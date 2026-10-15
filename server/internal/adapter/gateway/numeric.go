// Package gateway implements each service's Repository port against
// Postgres via pgx + Squirrel — the only place that translates between
// domain entities and database rows.
package gateway

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"jww/internal/domain"
)

// moneyToNumeric/numericToMoney and exchangeRateToNumeric/
// numericToExchangeRate convert between our domain.Money/ExchangeRate
// (fixed-point int64s) and pgx's NUMERIC wire type, going through each
// type's plain-text form — the one format both sides parse exactly, with
// no float64 round trip anywhere.
func moneyToNumeric(m domain.Money) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(m.String()); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("encoding %s as numeric: %w", m.String(), err)
	}
	return n, nil
}

func numericToMoney(n pgtype.Numeric) (domain.Money, error) {
	s, err := numericText(n)
	if err != nil {
		return 0, err
	}
	return domain.ParseMoney(s)
}

func exchangeRateToNumeric(r domain.ExchangeRate) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(r.String()); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("encoding %s as numeric: %w", r.String(), err)
	}
	return n, nil
}

func numericToExchangeRate(n pgtype.Numeric) (domain.ExchangeRate, error) {
	s, err := numericText(n)
	if err != nil {
		return 0, err
	}
	return domain.ParseExchangeRate(s)
}

func numericText(n pgtype.Numeric) (string, error) {
	if !n.Valid {
		return "", fmt.Errorf("numeric value is NULL")
	}
	value, err := n.Value()
	if err != nil {
		return "", fmt.Errorf("reading numeric value: %w", err)
	}
	s, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("unexpected numeric driver value %T", value)
	}
	return s, nil
}
