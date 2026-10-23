package openai

import (
	"fmt"

	"jww/internal/domain"
)

// Check verifies provider connectivity, credentials, and availability of the
// configured extraction model without creating a completion.
func (c *Client) Check(ctx domain.AppContext) error {
	if _, err := c.client.Models.Get(ctx.Context(), c.config.Model); err != nil {
		return fmt.Errorf("openai: checking model %q: %w", c.config.Model, err)
	}
	return nil
}
