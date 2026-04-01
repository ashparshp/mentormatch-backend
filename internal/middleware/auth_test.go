package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	secret := "test_secret"
	userID := "user-123"
	role := "mentor"

	// Helper to generate a valid token
	generateToken := func(id, role string, exp time.Duration) string {
		claims := jwt.MapClaims{
			"sub": id,
			"aud": "authenticated",
			"exp": time.Now().Add(exp).Unix(),
			"app_metadata": map[string]interface{}{
				"role": role,
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		s, _ := token.SignedString([]byte(secret))
		return s
	}

	tests := []struct {
		name          string
		token         string
		expectedStatus int
	}{
		{
			name:           "Valid Token",
			token:          generateToken(userID, role, time.Hour),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Expired Token",
			token:          generateToken(userID, role, -time.Hour),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Format",
			token:          "invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing Header",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Auth(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctxUserID := r.Context().Value(UserIDKey).(string)
				ctxUserRole := r.Context().Value(UserRoleKey).(string)
				
				assert.Equal(t, userID, ctxUserID)
				assert.Equal(t, role, ctxUserRole)
				
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tt.token))
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestRoleGuard(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		allowedRoles   []string
		expectedStatus int
	}{
		{
			name:           "Authorized Mentor",
			userRole:       "mentor",
			allowedRoles:   []string{"mentor", "admin"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized Student",
			userRole:       "student",
			allowedRoles:   []string{"mentor"},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RoleGuard(tt.allowedRoles...)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			rr := httptest.NewRecorder()

			// Inject user role directly into context
			ctx := req.Context()
			ctx = context.WithValue(ctx, UserRoleKey, tt.userRole)
			
			handler.ServeHTTP(rr, req.WithContext(ctx))

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
