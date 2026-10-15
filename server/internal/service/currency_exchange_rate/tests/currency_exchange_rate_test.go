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
	byPair      map[string]domain.CurrencyExchangeRate
	createCalls int
	createErr   error
	getErr      error
	listResult  []domain.CurrencyExchangeRate
	listErr     error
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
