package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/auth"
	"jww/internal/service/unit_of_work"
)

func newAppContext() domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(context.Background())
	return appCtx
}

type fakeUserRepo struct {
	byUsername   map[string]domain.User
	byID         map[uuid.UUID]domain.User
	getByIDErr   error
	getByUserErr error
}

func newFakeUserRepo(users ...domain.User) *fakeUserRepo {
	r := &fakeUserRepo{byUsername: map[string]domain.User{}, byID: map[uuid.UUID]domain.User{}}
	for _, u := range users {
		r.byUsername[u.Username] = u
		r.byID[u.ID] = u
	}
	return r
}

func (r *fakeUserRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.User, error) {
	if r.getByIDErr != nil {
		return nil, r.getByIDErr
	}
	if u, ok := r.byID[id]; ok {
		return &u, nil
	}
	return nil, nil
}

func (r *fakeUserRepo) GetByUsername(_ domain.AppContext, username string, _ ...unit_of_work.UnitOfWork) (*domain.User, error) {
	if r.getByUserErr != nil {
		return nil, r.getByUserErr
	}
	if u, ok := r.byUsername[username]; ok {
		return &u, nil
	}
	return nil, nil
}

type fakePasswordHasher struct{}

func (fakePasswordHasher) Hash(password string) (string, error) { return "hashed:" + password, nil }
func (fakePasswordHasher) Verify(password, hash string) bool    { return hash == "hashed:"+password }

type fakeTokenService struct {
	issueErr  error
	issuedFor uuid.UUID
	verifyFn  func(token string) (*uuid.UUID, error)
}

func (s *fakeTokenService) Issue(userID uuid.UUID) (string, error) {
	if s.issueErr != nil {
		return "", s.issueErr
	}
	s.issuedFor = userID
	return "token-for-" + userID.String(), nil
}

func (s *fakeTokenService) Verify(token string) (*uuid.UUID, error) {
	return s.verifyFn(token)
}

func TestService_Login_HappyPathIssuesToken(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "alice", PasswordHash: "hashed:s3cret"}
	repo := newFakeUserRepo(u)
	tokens := &fakeTokenService{}
	svc := auth.NewService(repo, fakePasswordHasher{}, tokens)

	token, err := svc.Login(newAppContext(), "alice", "s3cret")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, u.ID, tokens.issuedFor)
}

func TestService_Login_WrongPasswordIsGenericAuthenticationError(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "alice", PasswordHash: "hashed:s3cret"}
	repo := newFakeUserRepo(u)
	svc := auth.NewService(repo, fakePasswordHasher{}, &fakeTokenService{})

	_, err := svc.Login(newAppContext(), "alice", "wrong-password")
	assertIsAuthenticationError(t, err)
}

func TestService_Login_UnknownUsernameIsTheSameGenericError(t *testing.T) {
	repo := newFakeUserRepo() // no users
	svc := auth.NewService(repo, fakePasswordHasher{}, &fakeTokenService{})

	unknownUserErr := func() error {
		_, err := svc.Login(newAppContext(), "nobody", "whatever")
		return err
	}()
	wrongPasswordErr := func() error {
		u := domain.User{ID: uuid.New(), Username: "alice", PasswordHash: "hashed:s3cret"}
		repo := newFakeUserRepo(u)
		svc := auth.NewService(repo, fakePasswordHasher{}, &fakeTokenService{})
		_, err := svc.Login(newAppContext(), "alice", "wrong-password")
		return err
	}()

	// Anti-enumeration: an unknown username and a wrong password must be
	// indistinguishable to the caller.
	assertIsAuthenticationError(t, unknownUserErr)
	assert.Equal(t, wrongPasswordErr.Error(), unknownUserErr.Error())
}

func TestService_Login_RepoErrorPropagatesAsIs(t *testing.T) {
	repo := newFakeUserRepo()
	repo.getByUserErr = errors.New("boom")
	svc := auth.NewService(repo, fakePasswordHasher{}, &fakeTokenService{})

	_, err := svc.Login(newAppContext(), "alice", "s3cret")
	assert.ErrorIs(t, err, repo.getByUserErr, "a repo failure is not the same thing as bad credentials")
}

func TestService_Authenticate_HappyPath(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "alice"}
	repo := newFakeUserRepo(u)
	tokens := &fakeTokenService{verifyFn: func(string) (*uuid.UUID, error) { return &u.ID, nil }}
	svc := auth.NewService(repo, fakePasswordHasher{}, tokens)

	got, err := svc.Authenticate(newAppContext(), "a-valid-token")
	require.NoError(t, err)
	assert.Equal(t, u.ID, got.ID)
}

func TestService_Authenticate_InvalidTokenIsAuthenticationError(t *testing.T) {
	repo := newFakeUserRepo()
	tokens := &fakeTokenService{verifyFn: func(string) (*uuid.UUID, error) { return nil, errors.New("bad signature") }}
	svc := auth.NewService(repo, fakePasswordHasher{}, tokens)

	_, err := svc.Authenticate(newAppContext(), "garbage")
	assertIsAuthenticationError(t, err)
}

func TestService_Authenticate_ValidTokenForDeletedUserIsAuthenticationError(t *testing.T) {
	repo := newFakeUserRepo() // empty: the token's user no longer exists
	missingID := uuid.New()
	tokens := &fakeTokenService{verifyFn: func(string) (*uuid.UUID, error) { return &missingID, nil }}
	svc := auth.NewService(repo, fakePasswordHasher{}, tokens)

	_, err := svc.Authenticate(newAppContext(), "a-token-for-a-deleted-user")
	assertIsAuthenticationError(t, err)
}

func TestService_Authenticate_RepoErrorPropagatesAsIs(t *testing.T) {
	u := domain.User{ID: uuid.New()}
	repo := newFakeUserRepo(u)
	repo.getByIDErr = errors.New("boom")
	tokens := &fakeTokenService{verifyFn: func(string) (*uuid.UUID, error) { return &u.ID, nil }}
	svc := auth.NewService(repo, fakePasswordHasher{}, tokens)

	_, err := svc.Authenticate(newAppContext(), "a-valid-token")
	assert.ErrorIs(t, err, repo.getByIDErr, "an unexpected repo failure must not be collapsed into AuthenticationError")
}

func assertIsAuthenticationError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var domainErr domain.DomainError
	require.True(t, errors.As(err, &domainErr))
	assert.Equal(t, domain.ErrorKindUnauthorized, domainErr.Kind())
}
