package kernel

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"jww/internal/domain"
)

func TestWriteJSONWrapsDataEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusOK, map[string]string{"code": "USD"})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Data["code"] != "USD" {
		t.Fatalf("data.code = %q, want USD", body.Data["code"])
	}
}

func TestWriteJSONWrapsListInDataEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusOK, []string{"a", "b"})

	var body struct {
		Data []string `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body.Data) != 2 || body.Data[0] != "a" || body.Data[1] != "b" {
		t.Fatalf("data = %v, want [a b]", body.Data)
	}
}

func TestWriteErrorMapsKinds(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"not found", domain.NewCurrencyNotFoundError("USD"), http.StatusNotFound, CodeNotFound},
		{"authentication", domain.NewAuthenticationError(), http.StatusUnauthorized, CodeUnauthorized},
		{"unknown error is internal", errors.New("boom"), http.StatusInternalServerError, CodeInternal},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			WriteError(rec, tc.err)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			var body ErrorEnvelope
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Error.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", body.Error.Code, tc.wantCode)
			}
			if body.Error.Message == "" {
				t.Fatal("message is empty")
			}
		})
	}
}

func TestWriteErrorUnauthorizedSetsBearerChallenge(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, domain.NewAuthenticationError())

	if got := rec.Header().Get("WWW-Authenticate"); got != "Bearer" {
		t.Fatalf("WWW-Authenticate = %q, want Bearer", got)
	}
}

func TestWriteErrorInternalDoesNotLeakCause(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, errors.New("connection string leaked"))

	var body ErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error.Message != "internal server error" {
		t.Fatalf("message = %q, want %q", body.Error.Message, "internal server error")
	}
}

func TestWriteBadRequestUsesErrorEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteBadRequest(rec, errors.New("amount must not be negative"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var body ErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error.Code != CodeBadRequest {
		t.Fatalf("code = %q, want %q", body.Error.Code, CodeBadRequest)
	}
	if body.Error.Message != "amount must not be negative" {
		t.Fatalf("message = %q, want the original error", body.Error.Message)
	}
}

func TestWriteValidationErrorCarriesFieldDetails(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteValidationError(rec, "validation failed",
		FieldError{Field: "amount", Message: "must not be negative"},
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var body ErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body.Error.Details) != 1 {
		t.Fatalf("details = %v, want one entry", body.Error.Details)
	}
	if body.Error.Details[0].Field != "amount" {
		t.Fatalf("details[0].field = %q, want amount", body.Error.Details[0].Field)
	}
}
