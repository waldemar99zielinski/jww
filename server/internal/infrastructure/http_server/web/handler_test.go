package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWebHandler(t *testing.T) {
	staticDir := newTestWebBuild(t)
	handler := newHandler(Config{StaticDir: staticDir, APIBaseURL: "https://api.example.com/"})

	tests := []struct {
		name        string
		method      string
		path        string
		accept      string
		wantStatus  int
		wantBody    string
		wantCache   string
		wantContent string
		wantAllow   string
	}{
		{name: "root serves the app without content negotiation", method: http.MethodGet, path: "/", wantStatus: http.StatusOK, wantBody: "spa shell", wantCache: "no-cache"},
		{name: "entrypoint canonicalizes to root", method: http.MethodGet, path: "/index.html", wantStatus: http.StatusMovedPermanently, wantCache: "no-cache"},
		{name: "static asset is served", method: http.MethodGet, path: "/app.js", wantStatus: http.StatusOK, wantBody: "app asset"},
		{name: "nested static asset is served", method: http.MethodGet, path: "/assets/app-abc123.js", wantStatus: http.StatusOK, wantBody: "hashed asset"},
		{name: "static directories are not listed", method: http.MethodGet, path: "/assets/", accept: "text/html", wantStatus: http.StatusNotFound},
		{name: "SPA route falls back to index", method: http.MethodGet, path: "/transactions/42", accept: "text/html,application/xhtml+xml", wantStatus: http.StatusOK, wantBody: "spa shell", wantCache: "no-cache"},
		{name: "dotted SPA route falls back to index", method: http.MethodGet, path: "/users/jane.doe", accept: "text/html", wantStatus: http.StatusOK, wantBody: "spa shell", wantCache: "no-cache"},
		{name: "missing asset is not the SPA", method: http.MethodGet, path: "/missing.js", accept: "*/*", wantStatus: http.StatusNotFound},
		{name: "non-navigation unknown path is not the SPA", method: http.MethodGet, path: "/unknown", accept: "application/json", wantStatus: http.StatusNotFound},
		{name: "runtime config is generated", method: http.MethodGet, path: runtimeConfigPath, wantStatus: http.StatusOK, wantBody: `"apiBaseUrl":"https://api.example.com"`, wantCache: "no-store", wantContent: "application/javascript"},
		{name: "web endpoints reject writes", method: http.MethodPost, path: runtimeConfigPath, wantStatus: http.StatusMethodNotAllowed, wantAllow: "GET, HEAD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want it to contain %q", rec.Body.String(), tt.wantBody)
			}
			if got := rec.Header().Get("Cache-Control"); got != tt.wantCache {
				t.Errorf("Cache-Control = %q, want %q", got, tt.wantCache)
			}
			if tt.wantContent != "" && !strings.Contains(rec.Header().Get("Content-Type"), tt.wantContent) {
				t.Errorf("Content-Type = %q, want it to contain %q", rec.Header().Get("Content-Type"), tt.wantContent)
			}
			if got := rec.Header().Get("Allow"); got != tt.wantAllow {
				t.Errorf("Allow = %q, want %q", got, tt.wantAllow)
			}
		})
	}
}

func newTestWebBuild(t *testing.T) string {
	t.Helper()
	staticDir := t.TempDir()
	assetsDir := filepath.Join(staticDir, "assets")
	if err := os.Mkdir(assetsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		filepath.Join(staticDir, "index.html"):    "spa shell",
		filepath.Join(staticDir, "app.js"):        "app asset",
		filepath.Join(assetsDir, "app-abc123.js"): strings.Repeat("hashed asset;", 100),
	}
	for path, contents := range files {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return staticDir
}
