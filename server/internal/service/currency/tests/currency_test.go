package currency_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/currency"
	"jww/internal/service/unit_of_work"
)

func newAppContext() domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(context.Background())
	return appCtx
}

// fakeRepo implements currency.Repository.
type fakeRepo struct {
	byCode      map[string]domain.Currency
	createCalls int
	createErr   error
	getErr      error
	updateErr   error
	deleteErr   error
}

func newFakeRepo(currencies ...domain.Currency) *fakeRepo {
	byCode := make(map[string]domain.Currency, len(currencies))
	for _, c := range currencies {
		byCode[c.Code] = c
	}
	return &fakeRepo{byCode: byCode}
}

func (r *fakeRepo) Create(_ domain.AppContext, c domain.Currency, _ ...unit_of_work.UnitOfWork) error {
	r.createCalls++
	if r.createErr != nil {
		return r.createErr
	}
	r.byCode[c.Code] = c
	return nil
}

func (r *fakeRepo) GetByCode(_ domain.AppContext, code string, _ ...unit_of_work.UnitOfWork) (*domain.Currency, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if c, ok := r.byCode[code]; ok {
		return &c, nil
	}
	return nil, nil
}

func (r *fakeRepo) List(domain.AppContext, ...unit_of_work.UnitOfWork) ([]domain.Currency, error) {
	out := make([]domain.Currency, 0, len(r.byCode))
	for _, c := range r.byCode {
		out = append(out, c)
	}
	return out, nil
}

func (r *fakeRepo) Update(_ domain.AppContext, c domain.Currency, _ ...unit_of_work.UnitOfWork) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.byCode[c.Code] = c
	return nil
}

func (r *fakeRepo) Delete(_ domain.AppContext, c domain.Currency, _ ...unit_of_work.UnitOfWork) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.byCode, c.Code)
	return nil
}

// fakeUoW/fakeUoWManager record Restore/Finish calls so Update/Delete tests
// can assert the right transaction demarcation.
type fakeUoW struct {
	restored, finished bool
	restoreErr         error
}

func (u *fakeUoW) Restore(domain.AppContext) error { u.restored = true; return u.restoreErr }
func (u *fakeUoW) Finish(domain.AppContext) error  { u.finished = true; return nil }
func (u *fakeUoW) GetValue() interface{}           { return nil }

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

func TestService_Create_HappyPath(t *testing.T) {
	repo := newFakeRepo()
	svc := currency.NewService(repo, &fakeUoWManager{})

	got, err := svc.Create(newAppContext(), "USD", "US Dollar", true)
	require.NoError(t, err)
	assert.Equal(t, "USD", got.Code)
	assert.Equal(t, 1, repo.createCalls)
}

func TestService_Create_ValidationFailureNeverCallsRepo(t *testing.T) {
	repo := newFakeRepo()
	svc := currency.NewService(repo, &fakeUoWManager{})

	_, err := svc.Create(newAppContext(), "us", "US Dollar", true) // lowercase code
	assert.Error(t, err)
	assert.Equal(t, 0, repo.createCalls)
}

func TestService_GetByCode_NotFoundTranslatesNilNilToTypedError(t *testing.T) {
	repo := newFakeRepo()
	svc := currency.NewService(repo, &fakeUoWManager{})

	_, err := svc.GetByCode(newAppContext(), "ZZZ")
	require.Error(t, err)
	var notFound domain.DomainError
	require.True(t, errors.As(err, &notFound))
	assert.Equal(t, domain.ErrorKindNotFound, notFound.Kind())
}

func TestService_GetByCode_RepoErrorPropagates(t *testing.T) {
	repo := newFakeRepo()
	repo.getErr = errors.New("boom")
	svc := currency.NewService(repo, &fakeUoWManager{})

	_, err := svc.GetByCode(newAppContext(), "USD")
	assert.ErrorIs(t, err, repo.getErr)
}

func TestService_Update_HappyPathFinishesTransaction(t *testing.T) {
	repo := newFakeRepo(domain.Currency{Code: "USD", Name: "US Dollar"})
	manager := &fakeUoWManager{}
	svc := currency.NewService(repo, manager)

	got, err := svc.Update(newAppContext(), domain.Currency{Code: "USD", Name: "United States Dollar"})
	require.NoError(t, err)
	assert.Equal(t, "United States Dollar", got.Name)
	require.NotNil(t, manager.last())
	assert.True(t, manager.last().finished)
	assert.False(t, manager.last().restored)
}

func TestService_Update_NotFoundRestoresTransaction(t *testing.T) {
	repo := newFakeRepo()
	manager := &fakeUoWManager{}
	svc := currency.NewService(repo, manager)

	_, err := svc.Update(newAppContext(), domain.Currency{Code: "ZZZ", Name: "Ghost"})
	require.Error(t, err)
	var notFound domain.DomainError
	require.True(t, errors.As(err, &notFound))
	assert.Equal(t, domain.ErrorKindNotFound, notFound.Kind())
	require.NotNil(t, manager.last())
	assert.True(t, manager.last().restored)
	assert.False(t, manager.last().finished)
}

func TestService_Update_ValidationFailureNeverOpensTransaction(t *testing.T) {
	repo := newFakeRepo()
	manager := &fakeUoWManager{}
	svc := currency.NewService(repo, manager)

	_, err := svc.Update(newAppContext(), domain.Currency{Code: "us", Name: "Bad"})
	assert.Error(t, err)
	assert.Nil(t, manager.last(), "invalid input must be rejected before a transaction is ever opened")
}

func TestService_Update_RepoErrorRestoresTransaction(t *testing.T) {
	repo := newFakeRepo(domain.Currency{Code: "USD", Name: "US Dollar"})
	repo.updateErr = errors.New("boom")
	manager := &fakeUoWManager{}
	svc := currency.NewService(repo, manager)

	_, err := svc.Update(newAppContext(), domain.Currency{Code: "USD", Name: "New Name"})
	assert.ErrorIs(t, err, repo.updateErr)
	require.NotNil(t, manager.last())
	assert.True(t, manager.last().restored)
	assert.False(t, manager.last().finished)
}

func TestService_Delete_HappyPathFinishesTransaction(t *testing.T) {
	repo := newFakeRepo(domain.Currency{Code: "USD", Name: "US Dollar"})
	manager := &fakeUoWManager{}
	svc := currency.NewService(repo, manager)

	err := svc.Delete(newAppContext(), domain.Currency{Code: "USD"})
	require.NoError(t, err)
	assert.True(t, manager.last().finished)
	assert.False(t, manager.last().restored)
}

func TestService_Delete_NotFoundRestoresTransaction(t *testing.T) {
	repo := newFakeRepo()
	manager := &fakeUoWManager{}
	svc := currency.NewService(repo, manager)

	err := svc.Delete(newAppContext(), domain.Currency{Code: "ZZZ"})
	require.Error(t, err)
	assert.True(t, manager.last().restored)
	assert.False(t, manager.last().finished)
}
