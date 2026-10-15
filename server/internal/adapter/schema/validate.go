package schema

import "fmt"

// minLength enforces a minimum string length on request fields.
func minLength(field, value string, min int) error {
	if len(value) < min {
		return fmt.Errorf("%s must be at least %d character(s) long", field, min)
	}
	return nil
}

func exactLength(field, value string, n int) error {
	if len(value) != n {
		return fmt.Errorf("%s must be exactly %d characters long", field, n)
	}
	return nil
}

func nonNegative(field string, value int) error {
	if value < 0 {
		return fmt.Errorf("%s must not be negative", field)
	}
	return nil
}

func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
