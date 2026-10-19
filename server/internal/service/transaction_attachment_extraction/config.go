package transaction_attachment_extraction

// Config carries the absolute public base URL of THIS API, used to build the
// attachment-content link the vision provider fetches. Required: the provider
// must be able to reach it. It is loaded by cmd/server (not here) so this
// service package imports no configuration loader.
type Config struct {
	PublicBaseURL string `env:"PUBLIC_API_BASE_URL" bind:"required"`
}
