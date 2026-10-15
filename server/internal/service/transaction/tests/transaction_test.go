package transaction_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/transaction"
	"jww/internal/service/unit_of_work"
)

func newAppContext() domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(context.Background())
	return appCtx
}

var (
	testCurrency = domain.Currency{Code: "USD", Name: "US Dollar"}
	testCategory = domain.TransactionCategory{Type: domain.TransactionCategoryTypeExpense, Name: "Groceries", Active: true}
	testUser     = domain.User{Username: "alice"}
)

type fakeRepo struct {
	byID         map[int64]domain.Transaction
	nextID       int64
	createCalls  int
	createErr    error
	getErr       error
	updateErr    error
	deleteErr    error
	listResult   []domain.Transaction
	listErr      error
	lastListFilt *transaction.Filter
}

func newFakeRepo(transactions ...domain.Transaction) *fakeRepo {
	byID := make(map[int64]domain.Transaction, len(transactions))
	for _, tr := range transactions {
		byID[tr.ID] = tr
	}
	return &fakeRepo{byID: byID}
}

func (r *fakeRepo) Create(
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
	r.createCalls++
	if r.createErr != nil {
		return nil, r.createErr
	}
	r.nextID++
	tr := domain.Transaction{
		ID: r.nextID, Type: transactionType, Amount: amount, Currency: currency,
		Category: category, User: user, Source: source, Name: name,
		Description: description, CreatedAt: createdAt,
	}
	r.byID[tr.ID] = tr
	return &tr, nil
}

func (r *fakeRepo) GetByID(_ domain.AppContext, id int64, _ ...unit_of_work.UnitOfWork) (*domain.Transaction, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if tr, ok := r.byID[id]; ok {
		return &tr, nil
	}
	return nil, nil
}

func (r *fakeRepo) List(_ domain.AppContext, filter *transaction.Filter, _ ...unit_of_work.UnitOfWork) ([]domain.Transaction, error) {
	r.lastListFilt = filter
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.listResult, nil
}

func (r *fakeRepo) Update(_ domain.AppContext, tr domain.Transaction, _ ...unit_of_work.UnitOfWork) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.byID[tr.ID] = tr
	return nil
}

func (r *fakeRepo) Delete(_ domain.AppContext, tr domain.Transaction, _ ...unit_of_work.UnitOfWork) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.byID, tr.ID)
	return nil
}

type fakeUoW struct{ restored, finished bool }

func (u *fakeUoW) Restore(domain.AppContext) error { u.restored = true; return nil }
func (u *fakeUoW) Finish(domain.AppContext) error  { u.finished = true; return nil }
func (u *fakeUoW) GetValue() interface{}           { return nil }

type fakeUoWManager struct{ created []*fakeUoW }

func (m *fakeUoWManager) Create(domain.AppContext) (unit_of_work.UnitOfWork, error) {
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

func TestService_Create_HappyPath(t *testing.T) {
	repo := newFakeRepo()
	svc := transaction.NewService(repo, &fakeUoWManager{})

	got, err := svc.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(500),
		testCurrency, testCategory, testUser, domain.TransactionSourceManual, "Groceries run", "")
	require.NoError(t, err)
	assert.Equal(t, domain.Money(500), got.Amount)
	assert.Equal(t, 1, repo.createCalls)
}

func TestService_Create_ValidationFailureNeverCallsRepo(t *testing.T) {
	repo := newFakeRepo()
	svc := transaction.NewService(repo, &fakeUoWManager{})

	// category type (expense) doesn't match the transaction type (income).
	_, err := svc.Create(newAppContext(), domain.TransactionTypeIncome, domain.Money(500),
		testCurrency, testCategory, testUser, domain.TransactionSourceManual, "Mismatched", "")
	assert.Error(t, err)
	assert.Equal(t, 0, repo.createCalls)
}

func TestService_GetByID_NotFoundTranslatesNilNilToTypedError(t *testing.T) {
	repo := newFakeRepo()
	svc := transaction.NewService(repo, &fakeUoWManager{})

	_, err := svc.GetByID(newAppContext(), 999)
	require.Error(t, err)
	var notFound domain.DomainError
	require.True(t, errors.As(err, &notFound))
	assert.Equal(t, domain.ErrorKindNotFound, notFound.Kind())
}

func TestService_List_PassesFilterThrough(t *testing.T) {
	repo := newFakeRepo()
	want := []domain.Transaction{{ID: 1}}
	repo.listResult = want
	svc := transaction.NewService(repo, &fakeUoWManager{})

	filter := &transaction.Filter{Currency: &testCurrency}
	got, err := svc.List(newAppContext(), filter)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Same(t, filter, repo.lastListFilt)
}

func TestService_Update_HappyPathFinishesTransaction(t *testing.T) {
	existing := domain.Transaction{ID: 1, Type: domain.TransactionTypeExpense, Amount: 500, Category: testCategory, Name: "Old"}
	repo := newFakeRepo(existing)
	manager := &fakeUoWManager{}
	svc := transaction.NewService(repo, manager)

	updated := existing
	updated.Name = "New"
	got, err := svc.Update(newAppContext(), updated)
	require.NoError(t, err)
	assert.Equal(t, "New", got.Name)
	assert.True(t, manager.last().finished)
	assert.False(t, manager.last().restored)
}

func TestService_Update_NotFoundRestoresTransaction(t *testing.T) {
	repo := newFakeRepo()
	manager := &fakeUoWManager{}
	svc := transaction.NewService(repo, manager)

	ghost := domain.Transaction{ID: 999, Type: domain.TransactionTypeExpense, Amount: 500, Category: testCategory, Name: "Ghost"}
	_, err := svc.Update(newAppContext(), ghost)
	require.Error(t, err)
	assert.True(t, manager.last().restored)
	assert.False(t, manager.last().finished)
}

func TestService_Update_ValidationFailureNeverOpensTransaction(t *testing.T) {
	repo := newFakeRepo()
	manager := &fakeUoWManager{}
	svc := transaction.NewService(repo, manager)

	_, err := svc.Update(newAppContext(), domain.Transaction{ID: 1, Amount: -1})
	assert.Error(t, err)
	assert.Nil(t, manager.last())
}

func TestService_Update_RepoErrorRestoresTransaction(t *testing.T) {
	existing := domain.Transaction{ID: 1, Type: domain.TransactionTypeExpense, Amount: 500, Category: testCategory, Name: "Old"}
	repo := newFakeRepo(existing)
	repo.updateErr = errors.New("boom")
	manager := &fakeUoWManager{}
	svc := transaction.NewService(repo, manager)

	_, err := svc.Update(newAppContext(), existing)
	assert.ErrorIs(t, err, repo.updateErr)
	assert.True(t, manager.last().restored)
	assert.False(t, manager.last().finished)
}

func TestService_Delete_HappyPathFinishesTransaction(t *testing.T) {
	existing := domain.Transaction{ID: 1, Type: domain.TransactionTypeExpense, Amount: 500, Category: testCategory, Name: "Doomed"}
	repo := newFakeRepo(existing)
	manager := &fakeUoWManager{}
	svc := transaction.NewService(repo, manager)

	err := svc.Delete(newAppContext(), existing)
	require.NoError(t, err)
	assert.True(t, manager.last().finished)
	assert.False(t, manager.last().restored)
}

func TestService_Delete_NotFoundRestoresTransaction(t *testing.T) {
	repo := newFakeRepo()
	manager := &fakeUoWManager{}
	svc := transaction.NewService(repo, manager)

	err := svc.Delete(newAppContext(), domain.Transaction{ID: 999})
	require.Error(t, err)
	assert.True(t, manager.last().restored)
	assert.False(t, manager.last().finished)
}
