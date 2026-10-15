package main

import (
	"jww/internal/adapter/frankfurter"
	"jww/internal/adapter/security"
	"jww/internal/infrastructure/database"
	"jww/internal/infrastructure/http_server"
	"jww/internal/infrastructure/scheduler"
	"jww/utils/logger"
)

// Config aggregates every module's own configuration (each loaded from its
// own env vars via that module's LoadConfig) into one struct, populated in
// a single call at process startup.
type Config struct {
	Http        http_server.Config
	Database    database.Config
	Auth        security.Config
	Logger      logger.Config
	Scheduler   scheduler.Config
	Frankfurter frankfurter.Config
}

func LoadConfig() (Config, error) {
	httpConfig, err := http_server.LoadConfig()
	if err != nil {
		return Config{}, err
	}
	databaseConfig, err := database.LoadConfig()
	if err != nil {
		return Config{}, err
	}
	authConfig, err := security.LoadConfig()
	if err != nil {
		return Config{}, err
	}
	loggerConfig, err := logger.LoadConfig()
	if err != nil {
		return Config{}, err
	}
	schedulerConfig, err := scheduler.LoadConfig()
	if err != nil {
		return Config{}, err
	}
	frankfurterConfig, err := frankfurter.LoadConfig()
	if err != nil {
		return Config{}, err
	}
	return Config{
		Http:        httpConfig,
		Database:    databaseConfig,
		Auth:        authConfig,
		Logger:      loggerConfig,
		Scheduler:   schedulerConfig,
		Frankfurter: frankfurterConfig,
	}, nil
}
