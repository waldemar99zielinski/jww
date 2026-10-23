package http_server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	httpweb "jww/internal/infrastructure/http_server/web"
)

func TestApplicationHandlerDispatchesAPIAndWeb(t *testing.T) {
	api := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	web := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := newApplicationHandler(api, web)

	tests := []struct {
		path string
		want int
	}{
		{path: "/api", want: http.StatusTeapot},
		{path: "/api/v1/auth/me", want: http.StatusTeapot},
		{path: "/api-client", want: http.StatusNoContent},
		{path: "/transactions/42", want: http.StatusNoContent},
	}

	for _, tt := range tests {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if rec.Code != tt.want {
			t.Errorf("%s: status = %d, want %d", tt.path, rec.Code, tt.want)
		}
	}
}

func TestNewHandlerRejectsInvalidWebBuildBeforeRouting(t *testing.T) {
	webCfg := httpweb.Config{StaticDir: filepath.Join(t.TempDir(), "missing")}
	_, err := NewHandler(nil, webCfg)
	if err == nil || !strings.Contains(err.Error(), "WEB_STATIC_DIR") {
		t.Fatalf("NewHandler() error = %v, want invalid WEB_STATIC_DIR", err)
	}
}

func TestApplicationHandlerKeepsBrowserPolicyOnWebBranch(t *testing.T) {
	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("app"), 0o600); err != nil {
		t.Fatal(err)
	}
	webHandler, err := httpweb.New(httpweb.Config{StaticDir: staticDir})
	if err != nil {
		t.Fatal(err)
	}
	api := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := newApplicationHandler(api, webHandler)

	webRec := httptest.NewRecorder()
	webReq := httptest.NewRequest(http.MethodGet, "/", nil)
	webReq.Header.Set("Accept", "text/html")
	handler.ServeHTTP(webRec, webReq)
	if webRec.Header().Get("Content-Security-Policy") == "" {
		t.Error("web response must include browser security policy")
	}

	apiRec := httptest.NewRecorder()
	handler.ServeHTTP(apiRec, httptest.NewRequest(http.MethodGet, "/api/test", nil))
	if got := apiRec.Header().Get("Content-Security-Policy"); got != "" {
		t.Errorf("API response unexpectedly inherited web CSP %q", got)
	}
}
