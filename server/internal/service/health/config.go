package health

import "jww/utils/configuration"

// Config is read from READINESS_LATENCY_WARN_MS. LatencyWarnMS is the
// per-dependency readiness latency at/above which the service logs a WARN.
type Config struct {
	LatencyWarnMS int64 `env:"READINESS_LATENCY_WARN_MS" bind:"optional" default:"1000"`
}

func LoadConfig() (Config, error) {
	return configuration.Load[Config]()
}
