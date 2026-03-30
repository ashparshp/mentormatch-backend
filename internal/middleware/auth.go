package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ashparshp/mentormatch-backend/pkg/response"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
)

func Auth(jwtSecret string) func(http.Handler) http.Handler {
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
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				response.Error(w, http.StatusUnauthorized, "Invalid or expired token", "UNAUTHORIZED")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.Error(w, http.StatusUnauthorized, "Invalid token claims", "UNAUTHORIZED")
				return
			}

			// Supabase JWTs have 'aud' set to 'authenticated'
			if aud, ok := claims["aud"].(string); !ok || aud != "authenticated" {
				response.Error(w, http.StatusUnauthorized, "Invalid token audience", "UNAUTHORIZED")
				return
			}

			// Supabase JWTs have 'sub' as user ID
			userID, ok := claims["sub"].(string)
			if !ok {
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
