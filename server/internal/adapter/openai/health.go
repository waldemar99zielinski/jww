package openai

import (
	"fmt"

	"jww/internal/domain"
)

// Check verifies provider connectivity, credentials, and availability of the
// configured extraction model without creating a completion. Listing models is
// more widely implemented by OpenAI-compatible providers than the optional
// single-model endpoint, especially when model IDs contain a slash.
func (c *Client) Check(ctx domain.AppContext) error {
	models, err := c.client.Models.List(ctx.Context())
	if err != nil {
		return fmt.Errorf("openai: listing models: %w", err)
	}
	for _, model := range models.Data {
		if model.ID == c.config.Model {
			return nil
		}
	}
	return fmt.Errorf("openai: configured model %q is not available", c.config.Model)
}
