package http_server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"jww/internal/infrastructure/http_server/kernel"
)

func TestPublicVsProtectedRouting(t *testing.T) {
	router := NewRouter(&kernel.Deps{})

	cases := []struct {
		name   string
		method string
		path   string
		want   int
	}{
		{"login is public", http.MethodPost, "/api/v1/auth/login", http.StatusBadRequest},
		{"summary is protected", http.MethodGet, "/api/v1/summary", http.StatusUnauthorized},
		{"currencies protected", http.MethodGet, "/api/v1/currencies", http.StatusUnauthorized},
		{"me is protected", http.MethodGet, "/api/v1/auth/me", http.StatusUnauthorized},
	}

	for _, tc := range cases {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, nil)
		router.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Errorf("%s: got %d, want %d", tc.name, rec.Code, tc.want)
		}
	}
}
