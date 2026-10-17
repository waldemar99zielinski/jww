//go:build integration

package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/gateway"
	"jww/internal/domain"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
	"jww/internal/service/unit_of_work"
	"jww/internal/tests/pgtest"
)

// seedCurrencyPair creates two distinct currencies inside uow's transaction,
// a prerequisite every currency_exchange_rates row needs (both FKs).
func seedCurrencyPair(t *testing.T, appCtx domain.AppContext, uow unit_of_work.UnitOfWork, fromCode, toCode string) (domain.Currency, domain.Currency) {
	t.Helper()
	repo := gateway.NewPostgresCurrencyRepository(pool)
	from := domain.Currency{Code: fromCode, Name: fromCode + " currency", IsBase: false}
	to := domain.Currency{Code: toCode, Name: toCode + " currency", IsBase: false}
	require.NoError(t, repo.Create(appCtx, from, uow))
	require.NoError(t, repo.Create(appCtx, to, uow))
	return from, to
}

func TestPostgresCurrencyExchangeRateRepository_CreateGetByCurrencies_RoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	from, to := seedCurrencyPair(t, appCtx, uow, "ZRA", "ZRB")

	repo := gateway.NewPostgresCurrencyExchangeRateRepository(pool)
	recordedAt := time.Now().UTC().Truncate(time.Microsecond)
	// 1234567.8912 — well within numeric(18,4) and chosen to exercise every
	// decimal place; if this ever silently round-tripped through float64
	// instead of text, a value like this is where it would first drift.
	rate := domain.ExchangeRate(12345678912)

	rer := domain.CurrencyExchangeRate{FromCurrency: from, ToCurrency: to, Rate: rate, RecordedAt: recordedAt}
	require.NoError(t, repo.Create(appCtx, rer, uow))

	got, err := repo.GetByCurrencies(appCtx, from, to, recordedAt.Add(time.Second), uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, rate, got.Rate, "rate must round-trip exactly through NUMERIC, not drift via float64")
	assert.Equal(t, from.Code, got.FromCurrency.Code)
	assert.Equal(t, to.Code, got.ToCurrency.Code)
	assert.WithinDuration(t, recordedAt, got.RecordedAt, time.Millisecond)
}

func TestPostgresCurrencyExchangeRateRepository_GetByCurrencies_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	from, to := seedCurrencyPair(t, appCtx, uow, "ZRC", "ZRD")

	repo := gateway.NewPostgresCurrencyExchangeRateRepository(pool)
	got, err := repo.GetByCurrencies(appCtx, from, to, time.Now(), uow)
	require.NoError(t, err)
	assert.Nil(t, got, "no rate recorded yet must be (nil, nil)")
}

func TestPostgresCurrencyExchangeRateRepository_GetByCurrencies_PicksLatestAsOf(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	from, to := seedCurrencyPair(t, appCtx, uow, "ZRE", "ZRF")
	repo := gateway.NewPostgresCurrencyExchangeRateRepository(pool)

	older := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Microsecond)
	newer := time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Microsecond)
	require.NoError(t, repo.Create(appCtx, domain.CurrencyExchangeRate{FromCurrency: from, ToCurrency: to, Rate: domain.ExchangeRate(10000), RecordedAt: older}, uow))
	require.NoError(t, repo.Create(appCtx, domain.CurrencyExchangeRate{FromCurrency: from, ToCurrency: to, Rate: domain.ExchangeRate(20000), RecordedAt: newer}, uow))

	got, err := repo.GetByCurrencies(appCtx, from, to, time.Now(), uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, domain.ExchangeRate(20000), got.Rate, "must pick the most recent rate as of now")

	// Asking "as of" a time before the newer rate was recorded must still
	// return the older one.
	gotEarlier, err := repo.GetByCurrencies(appCtx, from, to, older.Add(time.Minute), uow)
	require.NoError(t, err)
	require.NotNil(t, gotEarlier)
	assert.Equal(t, domain.ExchangeRate(10000), gotEarlier.Rate)
}

func TestPostgresCurrencyExchangeRateRepository_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	from, to := seedCurrencyPair(t, appCtx, uow, "ZRG", "ZRH")
	otherFrom, otherTo := seedCurrencyPair(t, appCtx, uow, "ZRI", "ZRJ")
	repo := gateway.NewPostgresCurrencyExchangeRateRepository(pool)

	older := time.Now().UTC().Add(-time.Hour).Truncate(time.Microsecond)
	newer := time.Now().UTC().Truncate(time.Microsecond)
	require.NoError(t, repo.Create(appCtx, domain.CurrencyExchangeRate{FromCurrency: from, ToCurrency: to, Rate: 10000, RecordedAt: older}, uow))
	require.NoError(t, repo.Create(appCtx, domain.CurrencyExchangeRate{FromCurrency: otherFrom, ToCurrency: otherTo, Rate: 20000, RecordedAt: newer}, uow))

	t.Run("no filter", func(t *testing.T) {
		list, err := repo.List(appCtx, nil, uow)
		require.NoError(t, err)
		require.Len(t, list, 2)
	})

	t.Run("currency codes", func(t *testing.T) {
		list, err := repo.List(appCtx, &currency_exchange_rate.Filter{
			FromCurrencyCode: from.Code,
			ToCurrencyCode:   to.Code,
		}, uow)
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, from.Code, list[0].FromCurrency.Code)
		assert.Equal(t, to.Code, list[0].ToCurrency.Code)
	})

	t.Run("recorded range", func(t *testing.T) {
		fromTime := newer.Add(-time.Minute)
		list, err := repo.List(appCtx, &currency_exchange_rate.Filter{RecordedFrom: &fromTime}, uow)
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, otherFrom.Code, list[0].FromCurrency.Code)
	})
}

func TestPostgresCurrencyExchangeRateRepository_ConstraintViolations(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)

	t.Run("from and to must differ", func(t *testing.T) {
		t.Parallel()
		uow := pgtest.BeginRollback(t, pool, appCtx)
		from, _ := seedCurrencyPair(t, appCtx, uow, "ZRI", "ZRJ")
		repo := gateway.NewPostgresCurrencyExchangeRateRepository(pool)

		err := repo.Create(appCtx, domain.CurrencyExchangeRate{FromCurrency: from, ToCurrency: from, Rate: 10000, RecordedAt: time.Now()}, uow)
		assert.Error(t, err, "currency_exchange_rates_distinct_currencies should reject from == to")
	})

	t.Run("rate must be positive", func(t *testing.T) {
		t.Parallel()
		uow := pgtest.BeginRollback(t, pool, appCtx)
		from, to := seedCurrencyPair(t, appCtx, uow, "ZRK", "ZRL")
		repo := gateway.NewPostgresCurrencyExchangeRateRepository(pool)

		// domain.CurrencyExchangeRate.Validate() would reject Rate<=0 first;
		// this proves the DB's own CHECK still refuses it if that's ever
		// bypassed.
		err := repo.Create(appCtx, domain.CurrencyExchangeRate{FromCurrency: from, ToCurrency: to, Rate: 0, RecordedAt: time.Now()}, uow)
		assert.Error(t, err, "currency_exchange_rates_rate_positive should reject rate <= 0")
	})
}

func TestPostgresCurrencyExchangeRateRepository_UpsertMany_FreshBatchInserts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	seedCurrencyPair(t, appCtx, uow, "ZSM", "ZSN")

	repo := gateway.NewPostgresCurrencyExchangeRateRepository(pool)
	asOf := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)

	counts, err := repo.UpsertMany(appCtx, []domain.CurrencyExchangeRateFlatten{
		{FromCurrencyCode: "ZSM", ToCurrencyCode: "ZSN", Rate: 12345, RecordedAt: asOf},
		{FromCurrencyCode: "ZSN", ToCurrencyCode: "ZSM", Rate: 8100, RecordedAt: asOf},
	}, uow)
	require.NoError(t, err)
	assert.Equal(t, currency_exchange_rate.UpsertCounts{Inserted: 2}, counts)

	list, err := repo.List(appCtx, nil, uow)
	require.NoError(t, err)
	require.Len(t, list, 2)
}

func TestPostgresCurrencyExchangeRateRepository_UpsertMany_ReRunReplacesInPlace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	seedCurrencyPair(t, appCtx, uow, "ZSO", "ZSP")

	repo := gateway.NewPostgresCurrencyExchangeRateRepository(pool)
	asOf := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)

	_, err := repo.UpsertMany(appCtx, []domain.CurrencyExchangeRateFlatten{
		{FromCurrencyCode: "ZSO", ToCurrencyCode: "ZSP", Rate: 10000, RecordedAt: asOf},
	}, uow)
	require.NoError(t, err)

	counts, err := repo.UpsertMany(appCtx, []domain.CurrencyExchangeRateFlatten{
		{FromCurrencyCode: "ZSO", ToCurrencyCode: "ZSP", Rate: 20000, RecordedAt: asOf},
	}, uow)
	require.NoError(t, err)
	assert.Equal(t, currency_exchange_rate.UpsertCounts{Updated: 1}, counts, "a re-run replaces, never duplicates")

	list, err := repo.List(appCtx, nil, uow)
	require.NoError(t, err)
	require.Len(t, list, 1, "still exactly one row per pair per business day")
	assert.Equal(t, domain.ExchangeRate(20000), list[0].Rate)
}

func TestPostgresCurrencyExchangeRateRepository_UpsertMany_SplitsMixedCounts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	seedCurrencyPair(t, appCtx, uow, "ZQA", "ZQB")
	seedCurrencyPair(t, appCtx, uow, "ZQC", "ZQD")

	repo := gateway.NewPostgresCurrencyExchangeRateRepository(pool)
	asOf := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)

	_, err := repo.UpsertMany(appCtx, []domain.CurrencyExchangeRateFlatten{
		{FromCurrencyCode: "ZQA", ToCurrencyCode: "ZQB", Rate: 10000, RecordedAt: asOf},
	}, uow)
	require.NoError(t, err)

	counts, err := repo.UpsertMany(appCtx, []domain.CurrencyExchangeRateFlatten{
		{FromCurrencyCode: "ZQA", ToCurrencyCode: "ZQB", Rate: 11000, RecordedAt: asOf}, // existing
		{FromCurrencyCode: "ZQB", ToCurrencyCode: "ZQC", Rate: 12000, RecordedAt: asOf}, // new
	}, uow)
	require.NoError(t, err)
	assert.Equal(t, currency_exchange_rate.UpsertCounts{Inserted: 1, Updated: 1}, counts)
}
