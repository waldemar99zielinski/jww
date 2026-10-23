package web

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigValidateWebServing(t *testing.T) {
	validWebDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(validWebDir, "index.html"), []byte("app"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{name: "API-only mode"},
		{name: "valid same-origin web", cfg: Config{StaticDir: validWebDir}},
		{name: "valid external API origin", cfg: Config{StaticDir: validWebDir, APIBaseURL: "https://api.example.com"}},
		{name: "valid uppercase API scheme", cfg: Config{StaticDir: validWebDir, APIBaseURL: "HTTPS://api.example.com"}},
		{name: "API URL without web", cfg: Config{APIBaseURL: "https://api.example.com"}, wantErr: "requires WEB_STATIC_DIR"},
		{name: "missing directory", cfg: Config{StaticDir: filepath.Join(t.TempDir(), "missing")}, wantErr: "WEB_STATIC_DIR"},
		{name: "missing index", cfg: Config{StaticDir: t.TempDir()}, wantErr: "index.html"},
		{name: "relative API URL", cfg: Config{StaticDir: validWebDir, APIBaseURL: "/backend"}, wantErr: "absolute http(s) origin"},
		{name: "API URL with path", cfg: Config{StaticDir: validWebDir, APIBaseURL: "https://api.example.com/backend"}, wantErr: "without credentials, path"},
		{name: "API URL with credentials", cfg: Config{StaticDir: validWebDir, APIBaseURL: "https://user:pass@api.example.com"}, wantErr: "without credentials, path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
