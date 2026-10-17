package utils

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ParseInt64PathParam parses a named Chi path parameter as a signed 64-bit ID.
func ParseInt64PathParam(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, name), 10, 64)
}

// ParseUUIDPathParam parses a named Chi path parameter as a UUID.
func ParseUUIDPathParam(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, name))
}
