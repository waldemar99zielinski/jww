package logger

import "jww/utils/configuration"

type Config struct {
	Level  string `env:"LOG_LEVEL" bind:"optional" default:"info"`
	Pretty bool   `env:"LOG_PRETTY" bind:"optional" default:"false"`
}

func LoadConfig() (Config, error) {
	return configuration.Load[Config]()
}
