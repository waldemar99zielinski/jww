package webauthn_credential_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
	"jww/internal/service/webauthn_credential"
)

func newAppContext() domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(context.Background())
	return appCtx
}

type fakeRepo struct {
	byID            map[uuid.UUID]domain.WebAuthnCredential
	byCredentialID  map[string]uuid.UUID
	createCalls     int
	createErr       error
	getByIDErr      error
	getByCredErr    error
	getByUserErr    error
	updateSignErr   error
	deleteErr       error
	lastSignCount   uint32
	lastUsedAtCalls int
}

func newFakeRepo(credentials ...domain.WebAuthnCredential) *fakeRepo {
	byID := make(map[uuid.UUID]domain.WebAuthnCredential, len(credentials))
	byCredentialID := make(map[string]uuid.UUID, len(credentials))
	for _, c := range credentials {
		byID[c.ID] = c
		byCredentialID[string(c.CredentialID)] = c.ID
	}
	return &fakeRepo{byID: byID, byCredentialID: byCredentialID}
}

func (r *fakeRepo) Create(_ domain.AppContext, c domain.WebAuthnCredential, _ ...unit_of_work.UnitOfWork) error {
	r.createCalls++
	if r.createErr != nil {
		return r.createErr
	}
	r.byID[c.ID] = c
	r.byCredentialID[string(c.CredentialID)] = c.ID
	return nil
}

func (r *fakeRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.WebAuthnCredential, error) {
	if r.getByIDErr != nil {
		return nil, r.getByIDErr
	}
	if c, ok := r.byID[id]; ok {
		return &c, nil
	}
	return nil, nil
}

func (r *fakeRepo) GetByCredentialID(_ domain.AppContext, credentialID []byte, _ ...unit_of_work.UnitOfWork) (*domain.WebAuthnCredential, error) {
	if r.getByCredErr != nil {
		return nil, r.getByCredErr
	}
	if id, ok := r.byCredentialID[string(credentialID)]; ok {
		c := r.byID[id]
		return &c, nil
	}
	return nil, nil
}

func (r *fakeRepo) GetByUserID(_ domain.AppContext, userID uuid.UUID, _ ...unit_of_work.UnitOfWork) ([]domain.WebAuthnCredential, error) {
	if r.getByUserErr != nil {
		return nil, r.getByUserErr
	}
	var out []domain.WebAuthnCredential
	for _, c := range r.byID {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *fakeRepo) UpdateSignCount(_ domain.AppContext, id uuid.UUID, signCount uint32, lastUsedAt time.Time, _ ...unit_of_work.UnitOfWork) error {
	if r.updateSignErr != nil {
		return r.updateSignErr
	}
	c := r.byID[id]
	c.SignCount = signCount
	c.LastUsedAt = &lastUsedAt
	r.byID[id] = c
	r.lastSignCount = signCount
	r.lastUsedAtCalls++
	return nil
}

func (r *fakeRepo) Delete(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.byID, id)
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
	svc := webauthn_credential.NewService(repo, &fakeUoWManager{})
	userID := uuid.New()

	got, err := svc.Create(newAppContext(), userID, "iPhone", []byte("cred-1"), []byte("pub-1"), "none", []byte{0}, 0, []string{"internal"}, true)
	require.NoError(t, err)
	assert.Equal(t, userID, got.UserID)
	assert.NotEqual(t, uuid.Nil, got.ID)
	assert.False(t, got.CreatedAt.IsZero())
	assert.True(t, got.BackupEligible, "BackupEligible must be persisted, not left at its zero value")
	assert.Equal(t, 1, repo.createCalls)
}

func TestService_Create_ValidationFailureNeverCallsRepo(t *testing.T) {
	repo := newFakeRepo()
	svc := webauthn_credential.NewService(repo, &fakeUoWManager{})

	_, err := svc.Create(newAppContext(), uuid.New(), "", []byte("cred-1"), []byte("pub-1"), "none", []byte{0}, 0, nil, false)
	assert.Error(t, err)
	assert.Equal(t, 0, repo.createCalls)
}

func TestService_List_PassesUserIDThrough(t *testing.T) {
	userA, userB := uuid.New(), uuid.New()
	credA := domain.WebAuthnCredential{ID: uuid.New(), UserID: userA, Name: "A", CredentialID: []byte("a"), PublicKey: []byte("pa")}
	credB := domain.WebAuthnCredential{ID: uuid.New(), UserID: userB, Name: "B", CredentialID: []byte("b"), PublicKey: []byte("pb")}
	repo := newFakeRepo(credA, credB)
	svc := webauthn_credential.NewService(repo, &fakeUoWManager{})

	got, err := svc.List(newAppContext(), userA)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, credA.ID, got[0].ID)
}

func TestService_GetByCredentialID_MissStaysNilNilNotAnError(t *testing.T) {
	repo := newFakeRepo()
	svc := webauthn_credential.NewService(repo, &fakeUoWManager{})

	got, err := svc.GetByCredentialID(newAppContext(), []byte("does-not-exist"))
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestService_TouchSignCount_UpdatesCounterAndLastUsedAt(t *testing.T) {
	existing := domain.WebAuthnCredential{ID: uuid.New(), UserID: uuid.New(), Name: "iPhone", CredentialID: []byte("c"), PublicKey: []byte("p")}
	repo := newFakeRepo(existing)
	svc := webauthn_credential.NewService(repo, &fakeUoWManager{})

	require.NoError(t, svc.TouchSignCount(newAppContext(), existing.ID, 5, time.Now().UTC()))
	assert.Equal(t, uint32(5), repo.lastSignCount)
	assert.Equal(t, 1, repo.lastUsedAtCalls)
}

func TestService_Delete_HappyPathFinishesTransaction(t *testing.T) {
	userID := uuid.New()
	existing := domain.WebAuthnCredential{ID: uuid.New(), UserID: userID, Name: "iPhone", CredentialID: []byte("c"), PublicKey: []byte("p")}
	repo := newFakeRepo(existing)
	manager := &fakeUoWManager{}
	svc := webauthn_credential.NewService(repo, manager)

	err := svc.Delete(newAppContext(), userID, existing.ID)
	require.NoError(t, err)
	assert.True(t, manager.last().finished)
	assert.False(t, manager.last().restored)
}

func TestService_Delete_NotFoundRestoresTransaction(t *testing.T) {
	repo := newFakeRepo()
	manager := &fakeUoWManager{}
	svc := webauthn_credential.NewService(repo, manager)

	err := svc.Delete(newAppContext(), uuid.New(), uuid.New())
	require.Error(t, err)
	var notFound domain.DomainError
	require.True(t, errors.As(err, &notFound))
	assert.Equal(t, domain.ErrorKindNotFound, notFound.Kind())
	assert.True(t, manager.last().restored)
	assert.False(t, manager.last().finished)
}

func TestService_Delete_BelongsToAnotherUser_404sIdenticallyToNotFound(t *testing.T) {
	owner := uuid.New()
	attacker := uuid.New()
	existing := domain.WebAuthnCredential{ID: uuid.New(), UserID: owner, Name: "iPhone", CredentialID: []byte("c"), PublicKey: []byte("p")}
	repo := newFakeRepo(existing)
	manager := &fakeUoWManager{}
	svc := webauthn_credential.NewService(repo, manager)

	err := svc.Delete(newAppContext(), attacker, existing.ID)
	require.Error(t, err)
	var notFound domain.DomainError
	require.True(t, errors.As(err, &notFound))
	assert.Equal(t, domain.ErrorKindNotFound, notFound.Kind())
	assert.True(t, manager.last().restored)

	// The row must still exist — Delete never ran against the repo.
	_, stillThere := repo.byID[existing.ID]
	assert.True(t, stillThere)
}

func TestService_Delete_RepoErrorRestoresTransaction(t *testing.T) {
	userID := uuid.New()
	existing := domain.WebAuthnCredential{ID: uuid.New(), UserID: userID, Name: "iPhone", CredentialID: []byte("c"), PublicKey: []byte("p")}
	repo := newFakeRepo(existing)
	repo.deleteErr = errors.New("boom")
	manager := &fakeUoWManager{}
	svc := webauthn_credential.NewService(repo, manager)

	err := svc.Delete(newAppContext(), userID, existing.ID)
	assert.ErrorIs(t, err, repo.deleteErr)
	assert.True(t, manager.last().restored)
	assert.False(t, manager.last().finished)
}
