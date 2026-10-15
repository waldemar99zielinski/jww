package kernel

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"jww/internal/domain"
)

// Machine-readable error codes carried in the "error.code" field. They are the
// stable part of the error contract; "error.message" is for humans.
const (
	CodeBadRequest   = "bad_request"
	CodeUnauthorized = "unauthorized"
	CodeNotFound     = "not_found"
	CodeConflict     = "conflict"
	CodeInternal     = "internal_error"
)

// DecodeJSON decodes r's body into a T, rejecting unknown fields and
// trailing JSON values — a request body must contain exactly one object.
func DecodeJSON[T any](r *http.Request) (T, error) {
	var v T
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		var zero T
		return zero, err
	}
	if err := dec.Decode(new(struct{})); err != io.EOF {
		var zero T
		return zero, errors.New("request body must contain exactly one JSON value")
	}
	return v, nil
}

// Envelope is the standard success body: every 2xx JSON response wraps its
// payload in a "data" field, so clients always unwrap the same way. List
// endpoints put their array in Data like any other payload.
type Envelope struct {
	Data any `json:"data"`
}

// ErrorEnvelope is the standard non-2xx body.
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody is the machine- and human-readable payload of an error response.
// Details carries field-level validation failures when there are any.
type ErrorBody struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

// FieldError is a single field-level validation failure in ErrorBody.Details.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// WriteJSON writes v with the given status code as the standard success
// envelope {"data": v}. It is the single success path every handler uses.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	writeJSON(w, status, Envelope{Data: v})
}

// WriteOK writes a 200 response with v wrapped in the success envelope.
func WriteOK(w http.ResponseWriter, v any) { WriteJSON(w, http.StatusOK, v) }

// WriteCreated writes a 201 response with v wrapped in the success envelope.
func WriteCreated(w http.ResponseWriter, v any) { WriteJSON(w, http.StatusCreated, v) }

// WriteNoContent writes a 204 response with no body.
func WriteNoContent(w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }

// writeJSON encodes v verbatim (no envelope) with the given status code. It
// backs both the success and error writers; callers should use those.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false) // don't escape <, >, & in JSON output
		_ = enc.Encode(v)
	}
}

// WriteBadRequest reports a request-decoding or request-validation failure
// — the Go equivalent of FastAPI/Pydantic's automatic 422, collapsed to 400
// since there's no schema library generating field-level detail here.
func WriteBadRequest(w http.ResponseWriter, err error) {
	writeError(w, http.StatusBadRequest, CodeBadRequest, err.Error(), nil)
}

// WriteValidationError writes a 400 validation failure with optional
// field-level details, so the standard "details" array can be populated once
// a validator reports individual fields rather than a single message.
func WriteValidationError(w http.ResponseWriter, message string, details ...FieldError) {
	writeError(w, http.StatusBadRequest, CodeBadRequest, message, details)
}

// WriteError translates a service/domain-layer error into the standard HTTP
// error response — the one place allowed to know both "net/http" and
// "domain.DomainError" exist.
func WriteError(w http.ResponseWriter, err error) {
	var domainErr domain.DomainError
	if errors.As(err, &domainErr) {
		if domainErr.Kind() == domain.ErrorKindUnauthorized {
			w.Header().Set("WWW-Authenticate", "Bearer")
		}
		status, code := errorStatus(domainErr.Kind())
		writeError(w, status, code, err.Error(), nil)
		return
	}
	httpLog.Slog().Error("internal error", slog.String("error", err.Error()))
	writeError(w, http.StatusInternalServerError, CodeInternal, "internal server error", nil)
}

// writeError writes the standard error envelope with the given status/code.
func writeError(w http.ResponseWriter, status int, code, message string, details []FieldError) {
	writeJSON(w, status, ErrorEnvelope{
		Error: ErrorBody{Code: code, Message: message, Details: details},
	})
}

// errorStatus maps a domain error kind to an HTTP status and machine code.
// Unknown kinds fall through to 404, matching the historical default.
func errorStatus(kind domain.ErrorKind) (int, string) {
	switch kind {
	case domain.ErrorKindValidation:
		return http.StatusBadRequest, CodeBadRequest
	case domain.ErrorKindUnauthorized:
		return http.StatusUnauthorized, CodeUnauthorized
	case domain.ErrorKindConflict:
		return http.StatusConflict, CodeConflict
	case domain.ErrorKindInternal:
		return http.StatusInternalServerError, CodeInternal
	default:
		return http.StatusNotFound, CodeNotFound
	}
}
