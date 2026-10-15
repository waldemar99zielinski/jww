package gateway

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"jww/internal/domain"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
	"jww/internal/service/unit_of_work"
)

// PostgresCurrencyExchangeRateRepository implements
// service/currency_exchange_rate.Repository.
type PostgresCurrencyExchangeRateRepository struct {
	db DBTX
}

func NewPostgresCurrencyExchangeRateRepository(db DBTX) *PostgresCurrencyExchangeRateRepository {
	return &PostgresCurrencyExchangeRateRepository{db: db}
}

const exchangeRateSelectColumns = `
	er.rate, er.recorded_at,
	fc.code, fc.name, fc.is_base,
	tc.code, tc.name, tc.is_base`

const exchangeRateFrom = `
	currency_exchange_rates er
	JOIN currencies fc ON fc.code = er.from_currency_code
	JOIN currencies tc ON tc.code = er.to_currency_code`

func scanExchangeRateRow(row pgx.Row) (*domain.CurrencyExchangeRate, error) {
	var rateNumeric pgtype.Numeric
	var recordedAt time.Time
	var from, to domain.Currency
	err := row.Scan(&rateNumeric, &recordedAt, &from.Code, &from.Name, &from.IsBase, &to.Code, &to.Name, &to.IsBase)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rate, err := numericToExchangeRate(rateNumeric)
	if err != nil {
		return nil, err
	}
	return &domain.CurrencyExchangeRate{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         rate,
		RecordedAt:   recordedAt,
	}, nil
}

func (r *PostgresCurrencyExchangeRateRepository) Create(ctx domain.AppContext, exchangeRate domain.CurrencyExchangeRate, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	rateNumeric, err := exchangeRateToNumeric(exchangeRate.Rate)
	if err != nil {
		return err
	}
	sql, args, err := psql.Insert("currency_exchange_rates").
		Columns("from_currency_code", "to_currency_code", "rate", "recorded_at").
		Values(exchangeRate.FromCurrency.Code, exchangeRate.ToCurrency.Code, rateNumeric, exchangeRate.RecordedAt).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}

// UpsertMany writes the whole batch as one multi-row INSERT, replacing the
// rate of any row whose (from, to, recorded_at) already exists. RETURNING
// (xmax = 0) distinguishes fresh inserts from replaced rows so callers get
// inserted/updated counts back. Within one batch every row shares the same
// business-date recorded_at while each directed pair is unique, so the
// conflict target never repeats inside a single statement.
func (r *PostgresCurrencyExchangeRateRepository) UpsertMany(ctx domain.AppContext, rates []domain.CurrencyExchangeRateFlatten, uow ...unit_of_work.UnitOfWork) (currency_exchange_rate.UpsertCounts, error) {
	if len(rates) == 0 {
		return currency_exchange_rate.UpsertCounts{}, nil
	}
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return currency_exchange_rate.UpsertCounts{}, err
	}

	insert := psql.Insert("currency_exchange_rates").
		Columns("from_currency_code", "to_currency_code", "rate", "recorded_at")
	for _, rate := range rates {
		rateNumeric, err := exchangeRateToNumeric(rate.Rate)
		if err != nil {
			return currency_exchange_rate.UpsertCounts{}, err
		}
		insert = insert.Values(rate.FromCurrencyCode, rate.ToCurrencyCode, rateNumeric, rate.RecordedAt)
	}
	sql, args, err := insert.
		Suffix("ON CONFLICT (from_currency_code, to_currency_code, recorded_at) DO UPDATE SET rate = EXCLUDED.rate RETURNING (xmax = 0)").
		ToSql()
	if err != nil {
		return currency_exchange_rate.UpsertCounts{}, err
	}

	rows, err := db.Query(ctx.Context(), sql, args...)
	if err != nil {
		return currency_exchange_rate.UpsertCounts{}, err
	}
	defer rows.Close()

	var counts currency_exchange_rate.UpsertCounts
	for rows.Next() {
		var inserted bool
		if err := rows.Scan(&inserted); err != nil {
			return counts, err
		}
		if inserted {
			counts.Inserted++
		} else {
			counts.Updated++
		}
	}
	if err := rows.Err(); err != nil {
		return counts, err
	}
	return counts, nil
}

func (r *PostgresCurrencyExchangeRateRepository) GetByCurrencies(ctx domain.AppContext, from, to domain.Currency, asOf time.Time, uow ...unit_of_work.UnitOfWork) (*domain.CurrencyExchangeRate, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select(exchangeRateSelectColumns[1:]).
		From(exchangeRateFrom[1:]).
		Where("er.from_currency_code = ? AND er.to_currency_code = ? AND er.recorded_at <= ?", from.Code, to.Code, asOf).
		OrderBy("er.recorded_at DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}
	return scanExchangeRateRow(db.QueryRow(ctx.Context(), sql, args...))
}

func (r *PostgresCurrencyExchangeRateRepository) List(ctx domain.AppContext, uow ...unit_of_work.UnitOfWork) ([]domain.CurrencyExchangeRate, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select(exchangeRateSelectColumns[1:]).
		From(exchangeRateFrom[1:]).
		OrderBy("er.from_currency_code", "er.to_currency_code", "er.recorded_at").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(ctx.Context(), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []domain.CurrencyExchangeRate
	for rows.Next() {
		rate, err := scanExchangeRateRow(rows)
		if err != nil {
			return nil, err
		}
		if rate != nil {
			rates = append(rates, *rate)
		}
	}
	return rates, rows.Err()
}
