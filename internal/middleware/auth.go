package middleware

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ashparshp/mentormatch-backend/internal/config"
	"github.com/ashparshp/mentormatch-backend/pkg/response"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
)

// JWKS structures for manual parsing
type jwksKey struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	X   string `json:"x"`
	Y   string `json:"y"`
	Crv string `json:"crv"`
}

type jwksResponse struct {
	Keys []jwksKey `json:"keys"`
}

var (
	jwksCache     *ecdsa.PublicKey
	jwksCacheLock sync.RWMutex
	lastCacheTime time.Time
)

func fetchPublicKey(supabaseURL string) (*ecdsa.PublicKey, error) {
	jwksCacheLock.RLock()
	if jwksCache != nil && time.Since(lastCacheTime) < 1*time.Hour {
		defer jwksCacheLock.RUnlock()
		return jwksCache, nil
	}
	jwksCacheLock.RUnlock()

	jwksCacheLock.Lock()
	defer jwksCacheLock.Unlock()

	// Double check after acquiring lock
	if jwksCache != nil && time.Since(lastCacheTime) < 1*time.Hour {
		return jwksCache, nil
	}

	url := fmt.Sprintf("%s/auth/v1/.well-known/jwks.json", strings.TrimSuffix(supabaseURL, "/"))
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	for _, key := range jwks.Keys {
		if key.Alg == "ES256" && key.Kty == "EC" {
			xBytes, err := base64.RawURLEncoding.DecodeString(key.X)
			if err != nil {
				continue
			}
			yBytes, err := base64.RawURLEncoding.DecodeString(key.Y)
			if err != nil {
				continue
			}

			pubKey := &ecdsa.PublicKey{
				Curve: elliptic.P256(),
				X:     new(big.Int).SetBytes(xBytes),
				Y:     new(big.Int).SetBytes(yBytes),
			}
			jwksCache = pubKey
			lastCacheTime = time.Now()
			return pubKey, nil
		}
	}

	return nil, fmt.Errorf("no ES256 key found in JWKS")
}

func Auth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "Missing authorization header", "UNAUTHORIZED")
				return
			}

			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
				response.Error(w, http.StatusUnauthorized, "Invalid authorization header format", "UNAUTHORIZED")
				return
			}

			tokenString := bearerToken[1]
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Handle ES256 (Asymmetric)
				if _, ok := token.Method.(*jwt.SigningMethodECDSA); ok {
					if cfg.SupabaseURL == "" {
						return nil, fmt.Errorf("SUPABASE_URL not set, cannot verify ES256 token")
					}
					return fetchPublicKey(cfg.SupabaseURL)
				}

				// Handle HS256 (Symmetric/Legacy)
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
					return []byte(cfg.JWTSecret), nil
				}

				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			})

			if err != nil {
				tokenPreview := tokenString
				if len(tokenPreview) > 8 {
					tokenPreview = tokenPreview[:8] + "..."
				}
				slog.Error("JWT parsing failed", "error", err, "token_preview", tokenPreview)
				response.Error(w, http.StatusUnauthorized, "Invalid or expired token", "UNAUTHORIZED")
				return
			}

			if !token.Valid {
				slog.Error("JWT token is invalid")
				response.Error(w, http.StatusUnauthorized, "Invalid or expired token", "UNAUTHORIZED")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				slog.Error("Could not extract JWT claims")
				response.Error(w, http.StatusUnauthorized, "Invalid token claims", "UNAUTHORIZED")
				return
			}

			// Supabase JWTs have 'aud' set to 'authenticated'
			if aud, ok := claims["aud"].(string); !ok || aud != "authenticated" {
				slog.Error("Invalid JWT audience", "expected", "authenticated", "actual", claims["aud"])
				response.Error(w, http.StatusUnauthorized, "Invalid token audience", "UNAUTHORIZED")
				return
			}

			// Supabase JWTs have 'sub' as user ID
			userID, ok := claims["sub"].(string)
			if !ok {
				slog.Error("User ID (sub) not found in JWT claims")
				response.Error(w, http.StatusUnauthorized, "User ID not found in token", "UNAUTHORIZED")
				return
			}

			// Extract role from app_metadata (where Supabase stores it)
			role := "student" // default fallback
			if appMetadata, ok := claims["app_metadata"].(map[string]interface{}); ok {
				if r, ok := appMetadata["role"].(string); ok {
					role = r
				}
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserRoleKey, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RoleGuard middleware to restrict routes to specific roles
func RoleGuard(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				response.Error(w, http.StatusForbidden, "Access denied", "FORBIDDEN")
				return
			}

			authorized := false
			for _, role := range roles {
				if userRole == role {
					authorized = true
					break
				}
			}

			if !authorized {
				response.Error(w, http.StatusForbidden, "Insufficient permissions", "FORBIDDEN")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
