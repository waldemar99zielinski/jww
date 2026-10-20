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
		{"draft attachment upload is protected", http.MethodPost, "/api/v1/transaction-attachments", http.StatusUnauthorized},
		{"draft attachment get is protected", http.MethodGet, "/api/v1/transaction-attachments/4f4dbdf7-cdc5-451a-820d-e0180ceec831", http.StatusUnauthorized},
		{"draft extraction is protected", http.MethodPost, "/api/v1/transaction-attachments/4f4dbdf7-cdc5-451a-820d-e0180ceec831/extraction", http.StatusUnauthorized},
		{"nested attachment list is protected", http.MethodGet, "/api/v1/transactions/1/attachments", http.StatusUnauthorized},
		{"nested attachment upload is protected", http.MethodPost, "/api/v1/transactions/1/attachments", http.StatusUnauthorized},
		{"nested attachment attach is protected", http.MethodPut, "/api/v1/transactions/1/attachments/4f4dbdf7-cdc5-451a-820d-e0180ceec831", http.StatusUnauthorized},
		{"sub-transaction create is protected", http.MethodPost, "/api/v1/transactions/1/sub-transactions", http.StatusUnauthorized},
		{"sub-transaction update is protected", http.MethodPatch, "/api/v1/transactions/1/sub-transactions/4f4dbdf7-cdc5-451a-820d-e0180ceec831", http.StatusUnauthorized},
		{"sub-transaction delete is protected", http.MethodDelete, "/api/v1/transactions/1/sub-transactions/4f4dbdf7-cdc5-451a-820d-e0180ceec831", http.StatusUnauthorized},
		{"webauthn register begin is protected", http.MethodPost, "/api/v1/webauthn/credentials/register/begin", http.StatusUnauthorized},
		{"webauthn register finish is protected", http.MethodPost, "/api/v1/webauthn/credentials/register/finish", http.StatusUnauthorized},
		{"webauthn credentials list is protected", http.MethodGet, "/api/v1/webauthn/credentials", http.StatusUnauthorized},
		{"webauthn credential delete is protected", http.MethodDelete, "/api/v1/webauthn/credentials/4f4dbdf7-cdc5-451a-820d-e0180ceec831", http.StatusUnauthorized},
		{"webauthn login finish is public", http.MethodPost, "/api/v1/webauthn/login/finish", http.StatusBadRequest},
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
