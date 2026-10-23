package web

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"jww/utils/configuration"
)

// Config controls optional React application serving and its browser-visible
// API origin. An empty StaticDir disables web serving and leaves the process in
// API-only mode.
type Config struct {
	StaticDir  string `env:"WEB_STATIC_DIR" bind:"optional"`
	APIBaseURL string `env:"WEB_API_BASE_URL" bind:"optional"`
}

func LoadConfig() (Config, error) {
	cfg, err := configuration.Load[Config]()
	if err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("web configuration: %w", err)
	}
	return cfg, nil
}

func (c Config) Enabled() bool { return c.StaticDir != "" }

// Validate catches deployment errors before the process opens its listener.
func (c Config) Validate() error {
	if !c.Enabled() {
		if c.APIBaseURL != "" {
			return fmt.Errorf("WEB_API_BASE_URL requires WEB_STATIC_DIR")
		}
		return nil
	}

	info, err := os.Stat(c.StaticDir)
	if err != nil {
		return fmt.Errorf("WEB_STATIC_DIR %q: %w", c.StaticDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("WEB_STATIC_DIR %q is not a directory", c.StaticDir)
	}

	indexPath := filepath.Join(c.StaticDir, "index.html")
	indexInfo, err := os.Stat(indexPath)
	if err != nil {
		return fmt.Errorf("web entrypoint %q: %w", indexPath, err)
	}
	if !indexInfo.Mode().IsRegular() {
		return fmt.Errorf("web entrypoint %q is not a regular file", indexPath)
	}

	if c.APIBaseURL == "" {
		return nil
	}
	baseURL, err := url.Parse(c.APIBaseURL)
	if err != nil {
		return fmt.Errorf("WEB_API_BASE_URL: %w", err)
	}
	if (!strings.EqualFold(baseURL.Scheme, "http") && !strings.EqualFold(baseURL.Scheme, "https")) || baseURL.Host == "" {
		return fmt.Errorf("WEB_API_BASE_URL must be an absolute http(s) origin")
	}
	if baseURL.User != nil || (baseURL.Path != "" && baseURL.Path != "/") || baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return fmt.Errorf("WEB_API_BASE_URL must contain only scheme and host, without credentials, path, query, or fragment")
	}
	return nil
}
