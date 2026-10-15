package database

import "jww/utils/configuration"

// Config is read from DATABASE_URL — same env var goose and the dev
// scripts already expect.
type Config struct {
	URL string `env:"DATABASE_URL" bind:"required"`
}

func LoadConfig() (Config, error) {
	return configuration.Load[Config]()
}
