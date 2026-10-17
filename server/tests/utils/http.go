// Package utils holds helpers shared across the suites under server/tests/
// (currently just tests/e2e) — generic HTTP test driving, test deps wiring,
// and small OS-level test utilities. Nothing here is tagged: it's a plain
// importable package, like internal/tests/pgtest, not a _test.go file —
// Go can't import identifiers from another package's _test.go files, so
// anything meant to be reused across test packages has to live here instead.
package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
)

// Envelope mirrors kernel.Envelope/ErrorEnvelope loosely enough to decode
// either shape without importing the kernel package's unexported internals.
type Envelope[T any] struct {
	Data  T `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Client drives HTTP requests against a running server for black-box tests.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// NewClient builds a Client against baseURL. httpClient defaults to
// &http.Client{} when nil.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &Client{BaseURL: baseURL, HTTP: httpClient}
}

// Do sends method/path (relative to BaseURL) with an optional JSON body and
// bearer token, decodes the response envelope ({"data": ...} or {"error":
// ...}) into v, and returns the status code. Pass a real schema.*Request
// value as body (not a hand-rolled map) so domain.Money/ExchangeRate's
// custom JSON encoding is used automatically — those types serialize as a
// decimal-formatted bare number ("12.34"), not raw cents, and a map literal
// would silently get that wrong.
func (c *Client) Do(t testing.TB, method, path, token string, body any, v any) int {
	t.Helper()

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, c.BaseURL+path, reader)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.HTTP.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	DecodeBody(t, resp, v)
	return resp.StatusCode
}

// DecodeBody reads and JSON-decodes resp's body into v (a no-op if v is nil
// or the body is empty, e.g. a 204 No Content).
func DecodeBody(t testing.TB, resp *http.Response, v any) {
	t.Helper()
	if v == nil {
		return
	}
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	if len(raw) > 0 {
		require.NoErrorf(t, json.Unmarshal(raw, v), "decoding response body: %s", raw)
	}
}

// Login logs in as the given credentials and returns the access token,
// failing the test on any error.
func (c *Client) Login(t testing.TB, username, password string) string {
	t.Helper()
	var resp Envelope[schema.TokenResponse]
	status := c.Do(t, http.MethodPost, "/api/v1/auth/login", "", schema.LoginRequest{
		Username: username,
		Password: password,
	}, &resp)
	require.Equal(t, http.StatusOK, status, "login should succeed")
	require.NotEmpty(t, resp.Data.AccessToken)
	return resp.Data.AccessToken
}

// UploadFile POSTs one streaming multipart file part named "file" and returns
// the raw response so callers can assert status and decode the envelope.
func (c *Client) UploadFile(t testing.TB, path, token, filename, contentType string, content []byte) (*http.Response, []byte) {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename=%q`, filename))
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	part, err := writer.CreatePart(header)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.BaseURL+path, &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.HTTP.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp, raw
}

// Download performs a GET and returns the status, raw body, and headers — the
// one endpoint whose success body is bytes, not the JSON envelope.
func (c *Client) Download(t testing.TB, path, token string) (int, []byte, http.Header) {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.BaseURL+path, nil)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.HTTP.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, raw, resp.Header
}
