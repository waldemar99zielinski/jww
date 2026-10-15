package currency_exchange_rate_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
	"jww/internal/service/unit_of_work"
)

func newAppContext() domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(context.Background())
	return appCtx
}

var (
	usd = domain.Currency{Code: "USD", Name: "US Dollar"}
	eur = domain.Currency{Code: "EUR", Name: "Euro"}
)

type fakeRepo struct {
	byPair       map[string]domain.CurrencyExchangeRate
	createCalls  int
	createErr    error
	getErr       error
	listResult   []domain.CurrencyExchangeRate
	listErr      error
	upsertCalls  int
	upsertErr    error
	upserted     []domain.CurrencyExchangeRateFlatten
	upsertCounts currency_exchange_rate.UpsertCounts
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byPair: map[string]domain.CurrencyExchangeRate{}}
}

func (r *fakeRepo) Create(_ domain.AppContext, e domain.CurrencyExchangeRate, _ ...unit_of_work.UnitOfWork) error {
	r.createCalls++
	if r.createErr != nil {
		return r.createErr
	}
	r.byPair[e.FromCurrency.Code+"->"+e.ToCurrency.Code] = e
	return nil
}

func (r *fakeRepo) UpsertMany(_ domain.AppContext, rates []domain.CurrencyExchangeRateFlatten, _ ...unit_of_work.UnitOfWork) (currency_exchange_rate.UpsertCounts, error) {
	r.upsertCalls++
	if r.upsertErr != nil {
		return currency_exchange_rate.UpsertCounts{}, r.upsertErr
	}
	r.upserted = append(r.upserted, rates...)
	return r.upsertCounts, nil
}

func (r *fakeRepo) GetByCurrencies(_ domain.AppContext, from, to domain.Currency, _ time.Time, _ ...unit_of_work.UnitOfWork) (*domain.CurrencyExchangeRate, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if e, ok := r.byPair[from.Code+"->"+to.Code]; ok {
		return &e, nil
	}
	return nil, nil
}

func (r *fakeRepo) List(domain.AppContext, ...unit_of_work.UnitOfWork) ([]domain.CurrencyExchangeRate, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.listResult, nil
}

func TestService_Create_HappyPath(t *testing.T) {
	repo := newFakeRepo()
	svc := currency_exchange_rate.NewService(repo)

	got, err := svc.Create(newAppContext(), usd, eur, domain.ExchangeRate(9000))
	require.NoError(t, err)
	assert.Equal(t, domain.ExchangeRate(9000), got.Rate)
	assert.False(t, got.RecordedAt.IsZero())
	assert.Equal(t, 1, repo.createCalls)
}

func TestService_Create_ValidationFailureNeverCallsRepo(t *testing.T) {
	repo := newFakeRepo()
	svc := currency_exchange_rate.NewService(repo)

	_, err := svc.Create(newAppContext(), usd, usd, domain.ExchangeRate(9000)) // from == to
	assert.Error(t, err)
	assert.Equal(t, 0, repo.createCalls)

	_, err = svc.Create(newAppContext(), usd, eur, domain.ExchangeRate(-1)) // non-positive rate
	assert.Error(t, err)
	assert.Equal(t, 0, repo.createCalls)
}

func TestService_GetByCurrencies_NotFoundTranslatesNilNilToTypedError(t *testing.T) {
	repo := newFakeRepo()
	svc := currency_exchange_rate.NewService(repo)

	_, err := svc.GetByCurrencies(newAppContext(), usd, eur, time.Now())
	require.Error(t, err)
	var notFound domain.DomainError
	require.True(t, errors.As(err, &notFound))
	assert.Equal(t, domain.ErrorKindNotFound, notFound.Kind())
}

func TestService_GetByCurrencies_RepoErrorPropagates(t *testing.T) {
	repo := newFakeRepo()
	repo.getErr = errors.New("boom")
	svc := currency_exchange_rate.NewService(repo)

	_, err := svc.GetByCurrencies(newAppContext(), usd, eur, time.Now())
	assert.ErrorIs(t, err, repo.getErr)
}

func TestService_List_IsAPassthrough(t *testing.T) {
	want := []domain.CurrencyExchangeRate{{FromCurrency: usd, ToCurrency: eur, Rate: 9000, RecordedAt: time.Now()}}
	repo := newFakeRepo()
	repo.listResult = want
	svc := currency_exchange_rate.NewService(repo)

	got, err := svc.List(newAppContext())
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestService_UpsertMany_EmptyBatchIsANoOp(t *testing.T) {
	repo := newFakeRepo()
	svc := currency_exchange_rate.NewService(repo)

	counts, err := svc.UpsertMany(newAppContext(), nil)

	require.NoError(t, err)
	assert.Equal(t, currency_exchange_rate.UpsertCounts{}, counts)
	assert.Equal(t, 0, repo.upsertCalls, "an empty batch must not reach the repository")
}

func TestService_UpsertMany_ForwardsBatchAndReturnsRepoCounts(t *testing.T) {
	repo := newFakeRepo()
	repo.upsertCounts = currency_exchange_rate.UpsertCounts{Inserted: 4, Updated: 2}
	svc := currency_exchange_rate.NewService(repo)

	counts, err := svc.UpsertMany(newAppContext(), []domain.CurrencyExchangeRateFlatten{
		{FromCurrencyCode: "USD", ToCurrencyCode: "EUR", Rate: 9000, RecordedAt: time.Now().UTC()},
		{FromCurrencyCode: "EUR", ToCurrencyCode: "USD", Rate: 11000, RecordedAt: time.Now().UTC()},
	})

	require.NoError(t, err)
	assert.Equal(t, currency_exchange_rate.UpsertCounts{Inserted: 4, Updated: 2}, counts)
	assert.Equal(t, 1, repo.upsertCalls)
	assert.Len(t, repo.upserted, 2)
}

func TestService_UpsertMany_InvalidRowAbortsBeforeIORepoNeverCalled(t *testing.T) {
	t.Run("from == to", func(t *testing.T) {
		repo := newFakeRepo()
		svc := currency_exchange_rate.NewService(repo)

		_, err := svc.UpsertMany(newAppContext(), []domain.CurrencyExchangeRateFlatten{
			{FromCurrencyCode: "USD", ToCurrencyCode: "EUR", Rate: 9000, RecordedAt: time.Now().UTC()},
			{FromCurrencyCode: "USD", ToCurrencyCode: "USD", Rate: 9000, RecordedAt: time.Now().UTC()},
		})

		assert.Error(t, err)
		assert.Equal(t, 0, repo.upsertCalls)
	})

	t.Run("non-positive rate", func(t *testing.T) {
		repo := newFakeRepo()
		svc := currency_exchange_rate.NewService(repo)

		_, err := svc.UpsertMany(newAppContext(), []domain.CurrencyExchangeRateFlatten{
			{FromCurrencyCode: "USD", ToCurrencyCode: "EUR", Rate: 0, RecordedAt: time.Now().UTC()},
		})

		assert.Error(t, err)
		assert.Equal(t, 0, repo.upsertCalls)
	})
}

func TestService_UpsertMany_NormalizesRecordedAtToUTC(t *testing.T) {
	repo := newFakeRepo()
	svc := currency_exchange_rate.NewService(repo)

	zone := time.FixedZone("UTC+2", 2*60*60)
	_, err := svc.UpsertMany(newAppContext(), []domain.CurrencyExchangeRateFlatten{
		{FromCurrencyCode: "USD", ToCurrencyCode: "EUR", Rate: 9000, RecordedAt: time.Date(2026, 10, 14, 0, 0, 0, 0, zone)},
	})

	require.NoError(t, err)
	require.Len(t, repo.upserted, 1)
	assert.Equal(t, time.UTC, repo.upserted[0].RecordedAt.Location())
	assert.True(t, repo.upserted[0].RecordedAt.Equal(time.Date(2026, 10, 13, 22, 0, 0, 0, time.UTC)))
}
