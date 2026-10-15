package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
	"jww/internal/service/user"
)

func newAppContext() domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(context.Background())
	return appCtx
}

type fakeRepo struct {
	byID        map[uuid.UUID]domain.User
	createCalls int
	createErr   error
	getErr      error
	updateErr   error
	deleteErr   error
}

func newFakeRepo(users ...domain.User) *fakeRepo {
	byID := make(map[uuid.UUID]domain.User, len(users))
	for _, u := range users {
		byID[u.ID] = u
	}
	return &fakeRepo{byID: byID}
}

func (r *fakeRepo) Create(_ domain.AppContext, u domain.User, _ ...unit_of_work.UnitOfWork) error {
	r.createCalls++
	if r.createErr != nil {
		return r.createErr
	}
	r.byID[u.ID] = u
	return nil
}

func (r *fakeRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.User, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if u, ok := r.byID[id]; ok {
		return &u, nil
	}
	return nil, nil
}

func (r *fakeRepo) List(domain.AppContext, ...unit_of_work.UnitOfWork) ([]domain.User, error) {
	out := make([]domain.User, 0, len(r.byID))
	for _, u := range r.byID {
		out = append(out, u)
	}
	return out, nil
}

func (r *fakeRepo) Update(_ domain.AppContext, u domain.User, _ ...unit_of_work.UnitOfWork) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.byID[u.ID] = u
	return nil
}

func (r *fakeRepo) Delete(_ domain.AppContext, u domain.User, _ ...unit_of_work.UnitOfWork) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.byID, u.ID)
	return nil
}

// fakeHasher implements user.PasswordHasher without real hashing.
type fakeHasher struct{ hashErr error }

func (h fakeHasher) Hash(password string) (string, error) {
	if h.hashErr != nil {
		return "", h.hashErr
	}
	return "hashed:" + password, nil
}

func (h fakeHasher) Verify(password, hash string) bool { return hash == "hashed:"+password }

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

func TestService_Create_HappyPathHashesPasswordAndSetsDefaults(t *testing.T) {
	repo := newFakeRepo()
	svc := user.NewService(repo, fakeHasher{}, &fakeUoWManager{})

	got, err := svc.Create(newAppContext(), "alice", "s3cret", "Alice")
	require.NoError(t, err)
	assert.Equal(t, "alice", got.Username)
	assert.Equal(t, "hashed:s3cret", got.PasswordHash, "the raw password must never be stored")
	assert.NotEqual(t, uuid.Nil, got.ID)
	assert.False(t, got.CreatedAt.IsZero())
	assert.Equal(t, 1, repo.createCalls)
}

func TestService_Create_HasherErrorNeverCallsRepo(t *testing.T) {
	repo := newFakeRepo()
	svc := user.NewService(repo, fakeHasher{hashErr: errors.New("boom")}, &fakeUoWManager{})

	_, err := svc.Create(newAppContext(), "alice", "s3cret", "Alice")
	assert.Error(t, err)
	assert.Equal(t, 0, repo.createCalls)
}

func TestService_Create_ValidationFailureNeverCallsRepo(t *testing.T) {
	repo := newFakeRepo()
	svc := user.NewService(repo, fakeHasher{}, &fakeUoWManager{})

	_, err := svc.Create(newAppContext(), "", "s3cret", "Alice") // blank username
	assert.Error(t, err)
	assert.Equal(t, 0, repo.createCalls)
}

func TestService_GetByID_NotFoundTranslatesNilNilToTypedError(t *testing.T) {
	repo := newFakeRepo()
	svc := user.NewService(repo, fakeHasher{}, &fakeUoWManager{})

	_, err := svc.GetByID(newAppContext(), uuid.New())
	require.Error(t, err)
	var notFound domain.DomainError
	require.True(t, errors.As(err, &notFound))
	assert.Equal(t, domain.ErrorKindNotFound, notFound.Kind())
}

func TestService_Update_HappyPathFinishesTransaction(t *testing.T) {
	existing := domain.User{ID: uuid.New(), Username: "alice", PasswordHash: "h", DisplayName: "Alice"}
	repo := newFakeRepo(existing)
	manager := &fakeUoWManager{}
	svc := user.NewService(repo, fakeHasher{}, manager)

	updated := existing
	updated.DisplayName = "Alice Updated"
	got, err := svc.Update(newAppContext(), updated)
	require.NoError(t, err)
	assert.Equal(t, "Alice Updated", got.DisplayName)
	assert.True(t, manager.last().finished)
	assert.False(t, manager.last().restored)
}

func TestService_Update_NotFoundRestoresTransaction(t *testing.T) {
	repo := newFakeRepo()
	manager := &fakeUoWManager{}
	svc := user.NewService(repo, fakeHasher{}, manager)

	_, err := svc.Update(newAppContext(), domain.User{ID: uuid.New(), Username: "ghost", PasswordHash: "h", DisplayName: "Ghost"})
	require.Error(t, err)
	assert.True(t, manager.last().restored)
	assert.False(t, manager.last().finished)
}

func TestService_Update_ValidationFailureNeverOpensTransaction(t *testing.T) {
	repo := newFakeRepo()
	manager := &fakeUoWManager{}
	svc := user.NewService(repo, fakeHasher{}, manager)

	_, err := svc.Update(newAppContext(), domain.User{ID: uuid.New(), Username: ""})
	assert.Error(t, err)
	assert.Nil(t, manager.last())
}

func TestService_Update_RepoErrorRestoresTransaction(t *testing.T) {
	existing := domain.User{ID: uuid.New(), Username: "alice", PasswordHash: "h", DisplayName: "Alice"}
	repo := newFakeRepo(existing)
	repo.updateErr = errors.New("boom")
	manager := &fakeUoWManager{}
	svc := user.NewService(repo, fakeHasher{}, manager)

	_, err := svc.Update(newAppContext(), existing)
	assert.ErrorIs(t, err, repo.updateErr)
	assert.True(t, manager.last().restored)
	assert.False(t, manager.last().finished)
}

func TestService_Delete_HappyPathFinishesTransaction(t *testing.T) {
	existing := domain.User{ID: uuid.New(), Username: "alice", PasswordHash: "h", DisplayName: "Alice"}
	repo := newFakeRepo(existing)
	manager := &fakeUoWManager{}
	svc := user.NewService(repo, fakeHasher{}, manager)

	err := svc.Delete(newAppContext(), existing)
	require.NoError(t, err)
	assert.True(t, manager.last().finished)
	assert.False(t, manager.last().restored)
}

func TestService_Delete_NotFoundRestoresTransaction(t *testing.T) {
	repo := newFakeRepo()
	manager := &fakeUoWManager{}
	svc := user.NewService(repo, fakeHasher{}, manager)

	err := svc.Delete(newAppContext(), domain.User{ID: uuid.New()})
	require.Error(t, err)
	assert.True(t, manager.last().restored)
	assert.False(t, manager.last().finished)
}
