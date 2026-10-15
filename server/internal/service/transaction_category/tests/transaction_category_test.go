package transactioncategory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/transaction_category"
	"jww/internal/service/unit_of_work"
)

func newAppContext() domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(context.Background())
	return appCtx
}

type fakeRepo struct {
	byID        map[uuid.UUID]domain.TransactionCategory
	createCalls int
	createErr   error
	getErr      error
	updateErr   error
	deleteErr   error
}

func newFakeRepo(categories ...domain.TransactionCategory) *fakeRepo {
	byID := make(map[uuid.UUID]domain.TransactionCategory, len(categories))
	for _, c := range categories {
		byID[c.ID] = c
	}
	return &fakeRepo{byID: byID}
}

func (r *fakeRepo) Create(_ domain.AppContext, c domain.TransactionCategory, _ ...unit_of_work.UnitOfWork) error {
	r.createCalls++
	if r.createErr != nil {
		return r.createErr
	}
	r.byID[c.ID] = c
	return nil
}

func (r *fakeRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.TransactionCategory, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if c, ok := r.byID[id]; ok {
		return &c, nil
	}
	return nil, nil
}

func (r *fakeRepo) ListByType(_ domain.AppContext, categoryType domain.TransactionCategoryType, activeOnly bool, _ ...unit_of_work.UnitOfWork) ([]domain.TransactionCategory, error) {
	var out []domain.TransactionCategory
	for _, c := range r.byID {
		if c.Type != categoryType {
			continue
		}
		if activeOnly && !c.Active {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *fakeRepo) Update(_ domain.AppContext, c domain.TransactionCategory, _ ...unit_of_work.UnitOfWork) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.byID[c.ID] = c
	return nil
}

func (r *fakeRepo) Delete(_ domain.AppContext, c domain.TransactionCategory, _ ...unit_of_work.UnitOfWork) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.byID, c.ID)
	return nil
}

type fakeUoW struct {
	restored, finished bool
}

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
	svc := transaction_category.NewService(repo, &fakeUoWManager{})

	got, err := svc.Create(newAppContext(), domain.TransactionCategoryTypeExpense, "Groceries", 0)
	require.NoError(t, err)
	assert.Equal(t, "Groceries", got.Name)
	assert.True(t, got.Active, "new categories default to active")
	assert.NotEqual(t, uuid.Nil, got.ID)
	assert.Equal(t, 1, repo.createCalls)
}

func TestService_Create_ValidationFailureNeverCallsRepo(t *testing.T) {
	repo := newFakeRepo()
	svc := transaction_category.NewService(repo, &fakeUoWManager{})

	_, err := svc.Create(newAppContext(), domain.TransactionCategoryTypeExpense, "", -1)
	assert.Error(t, err)
	assert.Equal(t, 0, repo.createCalls)
}

func TestService_GetByID_NotFoundTranslatesNilNilToTypedError(t *testing.T) {
	repo := newFakeRepo()
	svc := transaction_category.NewService(repo, &fakeUoWManager{})

	_, err := svc.GetByID(newAppContext(), uuid.New())
	require.Error(t, err)
	var notFound domain.DomainError
	require.True(t, errors.As(err, &notFound))
	assert.Equal(t, domain.ErrorKindNotFound, notFound.Kind())
}

func TestService_ListByType_FiltersByTypeAndActive(t *testing.T) {
	active := domain.TransactionCategory{ID: uuid.New(), Type: domain.TransactionCategoryTypeExpense, Name: "Active", Active: true}
	inactive := domain.TransactionCategory{ID: uuid.New(), Type: domain.TransactionCategoryTypeExpense, Name: "Inactive", Active: false}
	income := domain.TransactionCategory{ID: uuid.New(), Type: domain.TransactionCategoryTypeIncome, Name: "Income", Active: true}
	repo := newFakeRepo(active, inactive, income)
	svc := transaction_category.NewService(repo, &fakeUoWManager{})

	got, err := svc.ListByType(newAppContext(), domain.TransactionCategoryTypeExpense, true)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, active.ID, got[0].ID)
}

func TestService_Update_HappyPathFinishesTransaction(t *testing.T) {
	existing := domain.TransactionCategory{ID: uuid.New(), Type: domain.TransactionCategoryTypeExpense, Name: "Old", Active: true}
	repo := newFakeRepo(existing)
	manager := &fakeUoWManager{}
	svc := transaction_category.NewService(repo, manager)

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
	svc := transaction_category.NewService(repo, manager)

	_, err := svc.Update(newAppContext(), domain.TransactionCategory{ID: uuid.New(), Name: "Ghost"})
	require.Error(t, err)
	assert.True(t, manager.last().restored)
	assert.False(t, manager.last().finished)
}

func TestService_Update_ValidationFailureNeverOpensTransaction(t *testing.T) {
	repo := newFakeRepo()
	manager := &fakeUoWManager{}
	svc := transaction_category.NewService(repo, manager)

	_, err := svc.Update(newAppContext(), domain.TransactionCategory{ID: uuid.New(), Name: ""})
	assert.Error(t, err)
	assert.Nil(t, manager.last())
}

func TestService_Update_RepoErrorRestoresTransaction(t *testing.T) {
	existing := domain.TransactionCategory{ID: uuid.New(), Name: "Old"}
	repo := newFakeRepo(existing)
	repo.updateErr = errors.New("boom")
	manager := &fakeUoWManager{}
	svc := transaction_category.NewService(repo, manager)

	updated := existing
	updated.Name = "New"
	_, err := svc.Update(newAppContext(), updated)
	assert.ErrorIs(t, err, repo.updateErr)
	assert.True(t, manager.last().restored)
	assert.False(t, manager.last().finished)
}

func TestService_Delete_HappyPathFinishesTransaction(t *testing.T) {
	existing := domain.TransactionCategory{ID: uuid.New(), Name: "Doomed"}
	repo := newFakeRepo(existing)
	manager := &fakeUoWManager{}
	svc := transaction_category.NewService(repo, manager)

	err := svc.Delete(newAppContext(), existing)
	require.NoError(t, err)
	assert.True(t, manager.last().finished)
	assert.False(t, manager.last().restored)
}

func TestService_Delete_NotFoundRestoresTransaction(t *testing.T) {
	repo := newFakeRepo()
	manager := &fakeUoWManager{}
	svc := transaction_category.NewService(repo, manager)

	err := svc.Delete(newAppContext(), domain.TransactionCategory{ID: uuid.New()})
	require.Error(t, err)
	assert.True(t, manager.last().restored)
	assert.False(t, manager.last().finished)
}
