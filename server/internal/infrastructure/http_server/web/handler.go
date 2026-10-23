package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const runtimeConfigPath = "/runtime-config.js"

// New validates cfg and builds a handler for the React application. It has no
// API dependency; the parent http_server package owns top-level dispatch.
func New(cfg Config) (http.Handler, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if !cfg.Enabled() {
		return nil, fmt.Errorf("web serving is disabled")
	}
	return securityHeaders(cfg)(newHandler(cfg)), nil
}

func newHandler(cfg Config) http.Handler {
	files := http.FileServer(http.Dir(cfg.StaticDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path == runtimeConfigPath {
			serveRuntimeConfig(w, cfg.APIBaseURL)
			return
		}
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			serveIndex(w, r, cfg.StaticDir)
			return
		}

		urlPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		requestedPath := filepath.Join(cfg.StaticDir, filepath.FromSlash(urlPath))
		info, err := os.Stat(requestedPath)
		if err == nil {
			if info.Mode().IsRegular() {
				files.ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
			return
		}
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			// Permission and I/O failures indicate a broken deployment; serving the
			// SPA shell would hide the failure behind a misleading 200 response.
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Restrict the fallback to document navigations. Returning index.html for
		// a missing module would produce a misleading 200 and a MIME-type error.
		if !acceptsHTML(r.Header.Get("Accept")) {
			http.NotFound(w, r)
			return
		}

		serveIndex(w, r, cfg.StaticDir)
	})
}

func serveIndex(w http.ResponseWriter, r *http.Request, staticDir string) {
	// A new deployment may point at different hashed assets, so the shell must
	// be revalidated instead of outliving the release that generated it.
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
}

func acceptsHTML(accept string) bool {
	for _, value := range strings.Split(accept, ",") {
		mediaType := strings.TrimSpace(strings.SplitN(value, ";", 2)[0])
		if mediaType == "text/html" || mediaType == "application/xhtml+xml" {
			return true
		}
	}
	return false
}

func serveRuntimeConfig(w http.ResponseWriter, apiBaseURL string) {
	config, err := json.Marshal(struct {
		APIBaseURL string `json:"apiBaseUrl"`
	}{APIBaseURL: strings.TrimRight(apiBaseURL, "/")})
	if err != nil {
		http.Error(w, "failed to encode runtime config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	// Runtime configuration may change when the same image is promoted between
	// environments and must never be reused from a previous container instance.
	w.Header().Set("Cache-Control", "no-store")
	_, _ = fmt.Fprintf(w, "window.__JWW_CONFIG__ = Object.freeze(%s);\n", config)
}
