package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ashparshp/mentormatch-backend/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	cfg := &config.Config{
		Port:      "8080",
		LogLevel:  "info",
		JWTSecret: "test_secret",
	}
	
	// We pass nil for db as health check doesn't touch it right now
	srv := NewServer(cfg, nil)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())
}
