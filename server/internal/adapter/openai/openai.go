// Package openai implements transaction_attachment_extraction.Extractor against
// the OpenAI Chat Completions wire format (POST {base_url}/chat/completions,
// vision content parts, response_format json_schema structured output). That
// format is spoken by OpenAI itself, Azure OpenAI, and most self-hosted /
// open-weight servers, so switching providers is a BaseURL + APIKey + model
// config change, never a code change.
//
// It is the only place that speaks HTTP to an external vision provider. Each
// extraction flow has its own file, prompt, schema, and result decoder. Sending
// a vision message carrying absolute content URLs is shared here (the provider
// fetches bytes itself; no attachment bytes pass through us). Adapters don't
// log: they return errors and let the service and its caller log the outcome.
package openai

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	sdk "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"

	"jww/internal/domain"
)

// Client talks to an OpenAI-compatible Chat Completions API.
type Client struct {
	config Config
	client sdk.Client
}

// NewClient builds the adapter. It takes no error: a bad config surfaces per
// extraction call (and config loading itself already fails fast in LoadConfig).
// The SDK client carries the API key, provider base URL, an HTTP client bounded
// by TimeoutSeconds, and no retries — so a hung backend fails within the timeout
// instead of retrying past it.
func NewClient(config Config) *Client {
	return &Client{
		config: config,
		client: sdk.NewClient(
			option.WithAPIKey(config.APIKey),
			option.WithBaseURL(config.BaseURL),
			option.WithHTTPClient(&http.Client{Timeout: time.Duration(config.TimeoutSeconds) * time.Second}),
			option.WithMaxRetries(0),
		),
	}
}

// visionParams builds the shared wire request for one flow: configured model,
// temperature 0, strict JSON schema, and one user message containing the flow
// instruction followed by one image_url part per supplied URL. URLs are sent
// verbatim and in order.
func (c *Client) visionParams(instruction, schemaName string, schema map[string]any, imageURLs []string) sdk.ChatCompletionNewParams {
	content := []sdk.ChatCompletionContentPartUnionParam{sdk.TextContentPart(instruction)}
	for _, imageURL := range imageURLs {
		content = append(content, sdk.ImageContentPart(sdk.ChatCompletionContentPartImageImageURLParam{URL: imageURL}))
	}

	return sdk.ChatCompletionNewParams{
		Model:       c.config.Model,
		Temperature: param.NewOpt(0.0),
		ResponseFormat: sdk.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
				JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   schemaName,
					Schema: schema,
					Strict: param.NewOpt(true),
				},
			},
		},
		Messages: []sdk.ChatCompletionMessageParamUnion{
			sdk.UserMessage(content),
		},
	}
}

// structuredContent performs one vision request and returns the structured
// completion content. Every failure — transport error, timeout, non-2xx
// (including 429 and 5xx), a content_filter finish, or empty content — returns
// a plain error. The caller maps it to the 503 kind.
func (c *Client) structuredContent(ctx domain.AppContext, params sdk.ChatCompletionNewParams) (string, error) {
	completion, err := c.client.Chat.Completions.New(ctx.Context(), params)
	if err != nil {
		return "", requestError(err)
	}
	if completion == nil || len(completion.Choices) == 0 {
		return "", errors.New("openai: response carried no choices")
	}

	choice := completion.Choices[0]
	if choice.FinishReason == "content_filter" {
		return "", errors.New("openai: completion was filtered")
	}
	content := strings.TrimSpace(choice.Message.Content)
	if content == "" {
		return "", errors.New("openai: completion carried no structured output")
	}
	return content, nil
}

// requestError retains the provider's parsed diagnostic fields for internal
// boundary logging. The SDK's Error() deliberately contains only the HTTP
// status, which otherwise turns every provider-side 400 into an unactionable
// "Bad Request". The service still maps this to the generic public 503 error,
// so these details never reach the client response.
func requestError(err error) error {
	var apiErr *sdk.Error
	if !errors.As(err, &apiErr) {
		return fmt.Errorf("openai: requesting extraction: %w", err)
	}
	return fmt.Errorf(
		"openai: requesting extraction: status=%d code=%q type=%q param=%q message=%q: %w",
		apiErr.StatusCode,
		apiErr.Code,
		apiErr.Type,
		apiErr.Param,
		apiErr.Message,
		err,
	)
}
