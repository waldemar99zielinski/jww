package domain

import (
	"fmt"
	"strings"
)

// Currency is a plain data holder — construction/mutation logic lives in
// the service layer, not here. Validate only checks the entity's own
// invariants.
type Currency struct {
	Code   string
	Name   string
	IsBase bool
}

func (c Currency) Validate() error {
	if len(c.Code) != 3 || !isUpperAlpha(c.Code) {
		return fmt.Errorf("currency code must be a 3-letter uppercase ISO 4217 code")
	}
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("currency name must not be blank")
	}
	return nil
}

func isUpperAlpha(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
