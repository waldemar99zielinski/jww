package domain

import (
	"fmt"
	"math"
	"math/bits"
	"strconv"
	"strings"
)

// Money is a monetary amount in minor units (cents): value == cents/100.
// Holds a Postgres numeric(12,2) amount column exactly. Ordinary int64
// comparisons and arithmetic (==, <, +, -) work directly on it; the only
// operation that needs help is converting between currencies, since that
// mixes in ExchangeRate's different scale — see Convert.
type Money int64

// ExchangeRate is a currency conversion rate scaled by 1e4: value ==
// units/1e4. Holds a Postgres numeric(18,4) rate column exactly.
type ExchangeRate int64

const exchangeRateScale = 10_000 // 1e4, ExchangeRate's implied scale

// Convert applies the exchange rate to an amount, rounding the result to
// the nearest cent (half away from zero). The multiply happens in 128 bits
// (via math/bits) before dividing back down by the rate's scale, so it
// doesn't overflow int64 the way a naive cents*units would for realistic
// amounts and rates.
func (r ExchangeRate) Convert(amount Money) Money {
	if amount == 0 || r == 0 {
		return 0
	}
	sign := int64(1)
	cents := uint64(amount)
	if amount < 0 {
		sign = -sign
		cents = uint64(-amount)
	}
	units := uint64(r)
	if r < 0 {
		sign = -sign
		units = uint64(-r)
	}

	hi, lo := bits.Mul64(cents, units)
	if hi >= exchangeRateScale {
		panic("domain: currency conversion overflows int64")
	}
	q, rem := bits.Div64(hi, lo, exchangeRateScale)
	if rem*2 >= exchangeRateScale {
		q++
	}
	if q > math.MaxInt64 {
		panic("domain: currency conversion overflows int64")
	}
	return Money(sign * int64(q))
}

// parseFixedPoint reads a plain decimal string (no exponential notation)
// into an integer scaled by 10^decimals, rounding any extra fractional
// digits half away from zero.
func parseFixedPoint(label, s string, decimals int) (int64, error) {
	orig := s
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("%s: empty string", label)
	}
	negative := false
	switch s[0] {
	case '+':
		s = s[1:]
	case '-':
		negative = true
		s = s[1:]
	}

	intPart, fracPart, hasFrac := strings.Cut(s, ".")
	if hasFrac && strings.ContainsAny(fracPart, "+-") {
		return 0, fmt.Errorf("%s: invalid decimal %q", label, orig)
	}
	roundUp := false
	if len(fracPart) > decimals {
		roundUp = fracPart[decimals] >= '5'
		fracPart = fracPart[:decimals]
	}
	for len(fracPart) < decimals {
		fracPart += "0"
	}
	digits := intPart + fracPart
	if digits == "" {
		return 0, fmt.Errorf("%s: invalid decimal %q", label, orig)
	}

	value, err := strconv.ParseInt(digits, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid decimal %q: %w", label, orig, err)
	}
	if roundUp {
		value++
	}
	if negative {
		value = -value
	}
	return value, nil
}

// formatFixedPoint renders value (scaled by 10^decimals) with exactly
// `decimals` fractional digits, e.g. formatFixedPoint(1234, 2) == "12.34".
func formatFixedPoint(value int64, decimals int) string {
	negative := value < 0
	digits := strconv.FormatInt(value, 10)
	if negative {
		digits = digits[1:]
	}
	for len(digits) <= decimals {
		digits = "0" + digits
	}
	intPart := digits[:len(digits)-decimals]
	fracPart := digits[len(digits)-decimals:]
	out := intPart + "." + fracPart
	if negative {
		out = "-" + out
	}
	return out
}

func jsonString(data []byte) string {
	s := strings.TrimSpace(string(data))
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	return s
}

// ParseMoney reads a plain decimal string such as "12.34" or "-7" as cents.
func ParseMoney(s string) (Money, error) {
	cents, err := parseFixedPoint("money", s, 2)
	if err != nil {
		return 0, err
	}
	return Money(cents), nil
}

// String renders the amount in plain notation with exactly 2 decimals,
// e.g. "12.34" or "-7.00".
func (m Money) String() string { return formatFixedPoint(int64(m), 2) }

// MarshalJSON emits the amount as a bare (unquoted) integer number of minor
// units (cents), e.g. 1234. The HTTP contract is cents-only; decimal digits are
// a display concern, never a wire representation.
func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(m), 10)), nil
}

// UnmarshalJSON accepts a bare integer or a quoted integer, both in minor units
// (cents). Decimal input is rejected so a client that accidentally sends major
// units cannot be silently accepted as a different amount.
func (m *Money) UnmarshalJSON(data []byte) error {
	s := jsonString(data)
	if s == "null" {
		*m = 0
		return nil
	}
	if strings.Contains(s, ".") {
		return fmt.Errorf("money must be an integer number of cents, got %q", s)
	}
	cents, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return fmt.Errorf("money must be an integer number of cents: %w", err)
	}
	*m = Money(cents)
	return nil
}

// ParseExchangeRate reads a plain decimal string such as "4.1234".
func ParseExchangeRate(s string) (ExchangeRate, error) {
	units, err := parseFixedPoint("exchange rate", s, 4)
	if err != nil {
		return 0, err
	}
	return ExchangeRate(units), nil
}

// String renders the rate in plain notation with exactly 4 decimals, e.g.
// "4.1234".
func (r ExchangeRate) String() string { return formatFixedPoint(int64(r), 4) }

// MarshalJSON emits the rate as a bare (unquoted) JSON number token,
// matching the numeric(18,4) column it round-trips through.
func (r ExchangeRate) MarshalJSON() ([]byte, error) { return []byte(r.String()), nil }

// UnmarshalJSON accepts either a bare JSON number or a quoted string.
func (r *ExchangeRate) UnmarshalJSON(data []byte) error {
	s := jsonString(data)
	if s == "null" {
		*r = 0
		return nil
	}
	parsed, err := ParseExchangeRate(s)
	if err != nil {
		return err
	}
	*r = parsed
	return nil
}
