package main

import (
	"jww/internal/adapter/frankfurter"
	"jww/internal/adapter/openai"
	jwtadapter "jww/internal/adapter/security/jwt"
	webauthnadapter "jww/internal/adapter/security/webauthn"
	"jww/internal/adapter/storage"
	"jww/internal/infrastructure/database"
	"jww/internal/infrastructure/http_server"
	httpweb "jww/internal/infrastructure/http_server/web"
	"jww/internal/infrastructure/scheduler"
	healthservice "jww/internal/service/health"
	transaction_attachment_extraction "jww/internal/service/transaction_attachment_extraction"
	"jww/utils/configuration"
	"jww/utils/logger"
)

// Config aggregates every module's own configuration (each loaded from its
// own env vars via that module's LoadConfig) into one struct, populated in
// a single call at process startup.
type Config struct {
	Http        http_server.Config
	Web         httpweb.Config
	Database    database.Config
	JWT         jwtadapter.Config
	WebAuthn    webauthnadapter.Config
	Logger      logger.Config
	Scheduler   scheduler.Config
	Health      healthservice.Config
	Frankfurter frankfurter.Config
	Storage     storage.Config
	OpenAI      openai.Config
	Extraction  transaction_attachment_extraction.Config
}

func LoadConfig() (Config, error) {
	httpConfig, err := http_server.LoadConfig()
	if err != nil {
		return Config{}, err
	}
	webConfig, err := httpweb.LoadConfig()
	if err != nil {
		return Config{}, err
	}
	databaseConfig, err := database.LoadConfig()
	if err != nil {
		return Config{}, err
	}
	jwtConfig, err := jwtadapter.LoadConfig()
	if err != nil {
		return Config{}, err
	}
	webauthnConfig, err := webauthnadapter.LoadConfig()
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
	healthConfig, err := healthservice.LoadConfig()
	if err != nil {
		return Config{}, err
	}
	frankfurterConfig, err := frankfurter.LoadConfig()
	if err != nil {
		return Config{}, err
	}
	storageConfig, err := storage.LoadConfig()
	if err != nil {
		return Config{}, err
	}
	openaiConfig, err := openai.LoadConfig()
	if err != nil {
		return Config{}, err
	}
	extractionConfig, err := configuration.Load[transaction_attachment_extraction.Config]()
	if err != nil {
		return Config{}, err
	}
	return Config{
		Http:        httpConfig,
		Web:         webConfig,
		Database:    databaseConfig,
		JWT:         jwtConfig,
		WebAuthn:    webauthnConfig,
		Logger:      loggerConfig,
		Scheduler:   schedulerConfig,
		Health:      healthConfig,
		Frankfurter: frankfurterConfig,
		Storage:     storageConfig,
		OpenAI:      openaiConfig,
		Extraction:  extractionConfig,
	}, nil
}
