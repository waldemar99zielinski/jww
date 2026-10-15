package frankfurter

import "jww/utils/configuration"

// Config is read from FRANKFURTER_BASE_URL / FRANKFURTER_TIMEOUT_SECONDS. No
// API key: Frankfurter is anonymous, so there is nothing secret to configure.
type Config struct {
	BaseURL        string `env:"FRANKFURTER_BASE_URL" bind:"optional" default:"https://api.frankfurter.dev"`
	TimeoutSeconds int64  `env:"FRANKFURTER_TIMEOUT_SECONDS" bind:"optional" default:"15"`
}

func LoadConfig() (Config, error) {
	return configuration.Load[Config]()
}
