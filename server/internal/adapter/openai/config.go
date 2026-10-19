package openai

import "jww/utils/configuration"

// Config is read from OPENAI_* environment variables. APIKey is the one secret:
// it is server-only and never reaches the client. BaseURL is the provider-swap
// knob — any server speaking the OpenAI Chat Completions wire format can back
// this adapter by pointing it elsewhere.
type Config struct {
	APIKey         string `env:"OPENAI_API_KEY" bind:"required"`
	BaseURL        string `env:"OPENAI_BASE_URL" bind:"optional" default:"https://api.openai.com/v1"`
	Model          string `env:"OPENAI_EXTRACTION_MODEL" bind:"optional" default:"gpt-4.1-mini"`
	TimeoutSeconds int64  `env:"OPENAI_TIMEOUT_SECONDS" bind:"optional" default:"300"`
}

func LoadConfig() (Config, error) {
	return configuration.Load[Config]()
}
