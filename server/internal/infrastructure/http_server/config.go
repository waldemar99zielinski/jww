package http_server

import (
	"fmt"
	"strings"

	"jww/utils/configuration"
)

// Config controls the HTTP listener and API middleware. The optional React
// application has its own injected web.Config.
type Config struct {
	Host         string `env:"HTTP_HOST" bind:"optional" default:"0.0.0.0"`
	Port         int    `env:"HTTP_PORT" bind:"optional" default:"8000"`
	CorsOrigins  string `env:"HTTP_CORS_ORIGINS" bind:"optional" default:"*"`
	MaxBodyBytes int64  `env:"HTTP_MAX_BODY_BYTES" bind:"optional" default:"1048576"`
}

func LoadConfig() (Config, error) {
	return configuration.Load[Config]()
}

// Addr is the host:port pair to pass to http.Server.
func (c Config) Addr() string { return fmt.Sprintf("%s:%d", c.Host, c.Port) }

// CorsOriginsList splits comma-separated HTTP_CORS_ORIGINS; "*" (the
// default) allows any origin — fine since auth is a bearer-style header,
// not cookies.
func (c Config) CorsOriginsList() []string {
	if c.CorsOrigins == "*" {
		return []string{"*"}
	}
	var origins []string
	for _, origin := range strings.Split(c.CorsOrigins, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
