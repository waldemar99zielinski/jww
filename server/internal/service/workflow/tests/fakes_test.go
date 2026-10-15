package workflow_test

import (
	"context"
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
	"jww/internal/service/transaction"
	"jww/internal/service/unit_of_work"
)

// newAppContext builds a minimal AppContext for tests that aren't driven by
// an *http.Request, mirroring scripts/bootstrap's non-HTTP construction.
func newAppContext() domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(context.Background())
	return appCtx
}

// fakeUoW is a hand-written unit_of_work.UnitOfWork that records whether
// Restore/Finish were called, so workflow tests can assert every error path
// rolls back and the happy path commits exactly once.
type fakeUoW struct {
	restored   bool
	finished   bool
	restoreErr error
	finishErr  error
}

func (u *fakeUoW) Restore(domain.AppContext) error {
	u.restored = true
	return u.restoreErr
}

func (u *fakeUoW) Finish(domain.AppContext) error {
	u.finished = true
	return u.finishErr
}

func (u *fakeUoW) GetValue() interface{} { return nil }

// fakeUoWManager hands out fakeUoW instances and remembers the last one
// created, for assertions.
type fakeUoWManager struct {
	createErr error
	created   []*fakeUoW
}

func (m *fakeUoWManager) Create(domain.AppContext) (unit_of_work.UnitOfWork, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	u := &fakeUoW{}
	m.created = append(m.created, u)
	return u, nil
}

func (m *fakeUoWManager) last() *fakeUoW {
	if len(m.created) == 0 {
		return nil
	}
	return m.created[len(m.created)-1]
}

// fakeCurrencyRepo implements currency.Repository.
type fakeCurrencyRepo struct {
	byCode map[string]domain.Currency
	getErr error
}

func newFakeCurrencyRepo(currencies ...domain.Currency) *fakeCurrencyRepo {
	byCode := make(map[string]domain.Currency, len(currencies))
	for _, c := range currencies {
		byCode[c.Code] = c
	}
	return &fakeCurrencyRepo{byCode: byCode}
}

func (r *fakeCurrencyRepo) Create(domain.AppContext, domain.Currency, ...unit_of_work.UnitOfWork) error {
	return nil
}

func (r *fakeCurrencyRepo) GetByCode(_ domain.AppContext, code string, _ ...unit_of_work.UnitOfWork) (*domain.Currency, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if c, ok := r.byCode[code]; ok {
		return &c, nil
	}
	return nil, nil
}

func (r *fakeCurrencyRepo) List(domain.AppContext, ...unit_of_work.UnitOfWork) ([]domain.Currency, error) {
	all := make([]domain.Currency, 0, len(r.byCode))
	for _, c := range r.byCode {
		all = append(all, c)
	}
	return all, nil
}

func (r *fakeCurrencyRepo) Update(domain.AppContext, domain.Currency, ...unit_of_work.UnitOfWork) error {
	return nil
}

func (r *fakeCurrencyRepo) Delete(domain.AppContext, domain.Currency, ...unit_of_work.UnitOfWork) error {
	return nil
}

// fakeExchangeRateRepo implements currency_exchange_rate.Repository.
type fakeExchangeRateRepo struct {
	byPair    map[string]domain.CurrencyExchangeRate // key: "FROM->TO"
	seeded    []domain.CurrencyExchangeRate
	getErr    error
	getCalls  int
	createErr error
	created   []domain.CurrencyExchangeRate
	listErr   error

	upsertErr    error
	upsertCalls  int
	upserted     []domain.CurrencyExchangeRateFlatten
	upsertCounts currency_exchange_rate.UpsertCounts
}

func newFakeExchangeRateRepo(rates ...domain.CurrencyExchangeRate) *fakeExchangeRateRepo {
	byPair := make(map[string]domain.CurrencyExchangeRate, len(rates))
	for _, r := range rates {
		byPair[r.FromCurrency.Code+"->"+r.ToCurrency.Code] = r
	}
	return &fakeExchangeRateRepo{byPair: byPair, seeded: rates}
}

func (r *fakeExchangeRateRepo) Create(_ domain.AppContext, e domain.CurrencyExchangeRate, _ ...unit_of_work.UnitOfWork) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.created = append(r.created, e)
	r.byPair[e.FromCurrency.Code+"->"+e.ToCurrency.Code] = e
	return nil
}

func (r *fakeExchangeRateRepo) UpsertMany(_ domain.AppContext, rates []domain.CurrencyExchangeRateFlatten, _ ...unit_of_work.UnitOfWork) (currency_exchange_rate.UpsertCounts, error) {
	r.upsertCalls++
	if r.upsertErr != nil {
		return currency_exchange_rate.UpsertCounts{}, r.upsertErr
	}
	r.upserted = append(r.upserted, rates...)
	return r.upsertCounts, nil
}

func (r *fakeExchangeRateRepo) GetByCurrencies(_ domain.AppContext, from, to domain.Currency, _ time.Time, _ ...unit_of_work.UnitOfWork) (*domain.CurrencyExchangeRate, error) {
	r.getCalls++
	if r.getErr != nil {
		return nil, r.getErr
	}
	if e, ok := r.byPair[from.Code+"->"+to.Code]; ok {
		return &e, nil
	}
	return nil, nil
}

func (r *fakeExchangeRateRepo) List(domain.AppContext, ...unit_of_work.UnitOfWork) ([]domain.CurrencyExchangeRate, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	all := make([]domain.CurrencyExchangeRate, 0, len(r.seeded)+len(r.created))
	all = append(all, r.seeded...)
	all = append(all, r.created...)
	return all, nil
}

// fakeRateSource implements currency_exchange_rate.RateSource. It records how
// it was called so tests can assert the codes/date the workflow asked for.
type fakeRateSource struct {
	fetch     currency_exchange_rate.RateFetch
	err       error
	calls     int
	lastCodes []string
	lastAsOf  time.Time
}

func (s *fakeRateSource) Rates(_ domain.AppContext, currencies []string, asOf time.Time) (currency_exchange_rate.RateFetch, error) {
	s.calls++
	s.lastCodes = currencies
	s.lastAsOf = asOf
	if s.err != nil {
		return currency_exchange_rate.RateFetch{}, s.err
	}
	return s.fetch, nil
}

// fakeCategoryRepo implements transactioncategory.Repository.
type fakeCategoryRepo struct {
	byID            map[uuid.UUID]domain.TransactionCategory
	getErr          error
	listByTypeCalls int
}

func newFakeCategoryRepo(categories ...domain.TransactionCategory) *fakeCategoryRepo {
	byID := make(map[uuid.UUID]domain.TransactionCategory, len(categories))
	for _, c := range categories {
		byID[c.ID] = c
	}
	return &fakeCategoryRepo{byID: byID}
}

func (r *fakeCategoryRepo) Create(domain.AppContext, domain.TransactionCategory, ...unit_of_work.UnitOfWork) error {
	return nil
}

func (r *fakeCategoryRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.TransactionCategory, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if c, ok := r.byID[id]; ok {
		return &c, nil
	}
	return nil, nil
}

func (r *fakeCategoryRepo) ListByType(domain.AppContext, domain.TransactionCategoryType, bool, ...unit_of_work.UnitOfWork) ([]domain.TransactionCategory, error) {
	r.listByTypeCalls++
	return nil, nil
}

func (r *fakeCategoryRepo) Update(domain.AppContext, domain.TransactionCategory, ...unit_of_work.UnitOfWork) error {
	return nil
}

func (r *fakeCategoryRepo) Delete(domain.AppContext, domain.TransactionCategory, ...unit_of_work.UnitOfWork) error {
	return nil
}

// fakeUserRepo implements user.Repository.
type fakeUserRepo struct {
	byID   map[uuid.UUID]domain.User
	getErr error
}

func newFakeUserRepo(users ...domain.User) *fakeUserRepo {
	byID := make(map[uuid.UUID]domain.User, len(users))
	for _, u := range users {
		byID[u.ID] = u
	}
	return &fakeUserRepo{byID: byID}
}

func (r *fakeUserRepo) Create(domain.AppContext, domain.User, ...unit_of_work.UnitOfWork) error {
	return nil
}

func (r *fakeUserRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.User, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if u, ok := r.byID[id]; ok {
		return &u, nil
	}
	return nil, nil
}

func (r *fakeUserRepo) List(domain.AppContext, ...unit_of_work.UnitOfWork) ([]domain.User, error) {
	return nil, nil
}

func (r *fakeUserRepo) Update(domain.AppContext, domain.User, ...unit_of_work.UnitOfWork) error {
	return nil
}

func (r *fakeUserRepo) Delete(domain.AppContext, domain.User, ...unit_of_work.UnitOfWork) error {
	return nil
}

// fakePasswordHasher implements user.PasswordHasher without doing any real
// hashing — this package only needs it to satisfy user.NewService.
type fakePasswordHasher struct{ hashErr error }

func (h fakePasswordHasher) Hash(password string) (string, error) {
	if h.hashErr != nil {
		return "", h.hashErr
	}
	return "hashed:" + password, nil
}

func (h fakePasswordHasher) Verify(password, hash string) bool {
	return hash == "hashed:"+password
}

// fakeTransactionRepo implements transaction.Repository.
type fakeTransactionRepo struct {
	createErr    error
	createCalled int
	nextID       int64

	listResult []domain.Transaction
	listErr    error
	listCalled int
}

func (r *fakeTransactionRepo) Create(
	_ domain.AppContext,
	transactionType domain.TransactionType,
	amount domain.Money,
	currency domain.Currency,
	category domain.TransactionCategory,
	user domain.User,
	source domain.TransactionSource,
	createdAt time.Time,
	name string,
	description string,
	_ ...unit_of_work.UnitOfWork,
) (*domain.Transaction, error) {
	r.createCalled++
	if r.createErr != nil {
		return nil, r.createErr
	}
	r.nextID++
	return &domain.Transaction{
		ID:          r.nextID,
		Type:        transactionType,
		Amount:      amount,
		Currency:    currency,
		Category:    category,
		User:        user,
		Source:      source,
		Name:        name,
		Description: description,
		CreatedAt:   createdAt,
	}, nil
}

func (r *fakeTransactionRepo) GetByID(domain.AppContext, int64, ...unit_of_work.UnitOfWork) (*domain.Transaction, error) {
	return nil, nil
}

func (r *fakeTransactionRepo) List(_ domain.AppContext, _ *transaction.Filter, _ ...unit_of_work.UnitOfWork) ([]domain.Transaction, error) {
	r.listCalled++
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.listResult, nil
}

func (r *fakeTransactionRepo) Update(domain.AppContext, domain.Transaction, ...unit_of_work.UnitOfWork) error {
	return nil
}

func (r *fakeTransactionRepo) Delete(domain.AppContext, domain.Transaction, ...unit_of_work.UnitOfWork) error {
	return nil
}
