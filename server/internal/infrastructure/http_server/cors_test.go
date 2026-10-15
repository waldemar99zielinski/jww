package http_server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"jww/internal/infrastructure/http_server/kernel"
)

// TestCORSHeaders guards against a regression where AllowedMethods held the
// literal "*": go-chi/cors does not expand it, so no method matched and the
// middleware returned before writing Access-Control-Allow-Origin.
func TestCORSHeaders(t *testing.T) {
	router := NewRouter(&kernel.Deps{CorsOrigins: []string{"*"}})

	t.Run("actual request", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/summary", nil)
		req.Header.Set("Origin", "http://localhost:3000")

		router.ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "*")
		}
	})

	t.Run("preflight request", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")

		router.ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "*")
		}
		if got := rec.Header().Get("Access-Control-Allow-Methods"); got != http.MethodPost {
			t.Errorf("Access-Control-Allow-Methods = %q, want %q", got, http.MethodPost)
		}
	})
}
