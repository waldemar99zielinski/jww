package frankfurter

import (
	"fmt"
	"net/http"

	"jww/internal/domain"
)

// Check verifies that the configured Frankfurter-compatible provider is
// reachable without fetching or persisting an exchange-rate matrix.
func (c *Client) Check(ctx domain.AppContext) error {
	req, err := http.NewRequestWithContext(ctx.Context(), http.MethodGet, c.baseURL+"/v2/currencies", nil)
	if err != nil {
		return fmt.Errorf("frankfurter: building health request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("frankfurter: health request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("frankfurter: health request: unexpected status %d", resp.StatusCode)
	}
	return nil
}
