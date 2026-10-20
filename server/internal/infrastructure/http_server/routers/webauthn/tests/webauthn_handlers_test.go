package webauthn_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/adapter/security"
	jwtadapter "jww/internal/adapter/security/jwt"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	webauthnrouter "jww/internal/infrastructure/http_server/routers/webauthn"
	"jww/internal/service/auth"
	"jww/internal/service/unit_of_work"
	"jww/internal/service/user"
	"jww/internal/service/webauthn_credential"
	"jww/internal/service/workflow"
)

// fakeUserRepo implements user.Repository (and auth.UserRepository).
type fakeUserRepo struct {
	byID map[uuid.UUID]domain.User
}

func (r *fakeUserRepo) Create(domain.AppContext, domain.User, ...unit_of_work.UnitOfWork) error {
	return nil
}
func (r *fakeUserRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.User, error) {
	if u, ok := r.byID[id]; ok {
		return &u, nil
	}
	return nil, nil
}
func (r *fakeUserRepo) GetByUsername(_ domain.AppContext, username string, _ ...unit_of_work.UnitOfWork) (*domain.User, error) {
	for _, u := range r.byID {
		if u.Username == username {
			return &u, nil
		}
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

// fakeCredentialRepo implements webauthn_credential.Repository.
type fakeCredentialRepo struct {
	byID           map[uuid.UUID]domain.WebAuthnCredential
	byCredentialID map[string]uuid.UUID
}

func newFakeCredentialRepo(credentials ...domain.WebAuthnCredential) *fakeCredentialRepo {
	byID := make(map[uuid.UUID]domain.WebAuthnCredential, len(credentials))
	byCredentialID := make(map[string]uuid.UUID, len(credentials))
	for _, c := range credentials {
		byID[c.ID] = c
		byCredentialID[string(c.CredentialID)] = c.ID
	}
	return &fakeCredentialRepo{byID: byID, byCredentialID: byCredentialID}
}

func (r *fakeCredentialRepo) Create(_ domain.AppContext, c domain.WebAuthnCredential, _ ...unit_of_work.UnitOfWork) error {
	r.byID[c.ID] = c
	r.byCredentialID[string(c.CredentialID)] = c.ID
	return nil
}
func (r *fakeCredentialRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.WebAuthnCredential, error) {
	if c, ok := r.byID[id]; ok {
		return &c, nil
	}
	return nil, nil
}
func (r *fakeCredentialRepo) GetByCredentialID(_ domain.AppContext, credentialID []byte, _ ...unit_of_work.UnitOfWork) (*domain.WebAuthnCredential, error) {
	if id, ok := r.byCredentialID[string(credentialID)]; ok {
		c := r.byID[id]
		return &c, nil
	}
	return nil, nil
}
func (r *fakeCredentialRepo) GetByUserID(_ domain.AppContext, userID uuid.UUID, _ ...unit_of_work.UnitOfWork) ([]domain.WebAuthnCredential, error) {
	var out []domain.WebAuthnCredential
	for _, c := range r.byID {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, nil
}
func (r *fakeCredentialRepo) UpdateSignCount(_ domain.AppContext, id uuid.UUID, signCount uint32, _ time.Time, _ ...unit_of_work.UnitOfWork) error {
	return nil
}
func (r *fakeCredentialRepo) Delete(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) error {
	delete(r.byID, id)
	return nil
}

type fakeUoW struct{}

func (fakeUoW) Restore(domain.AppContext) error { return nil }
func (fakeUoW) Finish(domain.AppContext) error  { return nil }
func (fakeUoW) GetValue() interface{}           { return nil }

type fakeUoWManager struct{}

func (fakeUoWManager) Create(domain.AppContext) (unit_of_work.UnitOfWork, error) {
	return fakeUoW{}, nil
}

// fakeCeremony implements webauthn_credential.Ceremony.
type fakeCeremony struct {
	beginRegOptions domain.WebAuthnCeremonyOptions
	beginRegSession domain.WebAuthnCeremonySession

	beginLoginOptions domain.WebAuthnCeremonyOptions
	beginLoginSession domain.WebAuthnCeremonySession

	loginRawID                                []byte
	finishLoginCredentialID                   []byte
	finishLoginSignCount                      uint32
	finishRegCredentialID, finishRegPublicKey []byte
	finishRegAttestationType                  string
}

func (c *fakeCeremony) BeginRegistration(domain.User, []domain.WebAuthnCredential) (domain.WebAuthnCeremonyOptions, domain.WebAuthnCeremonySession, error) {
	return c.beginRegOptions, c.beginRegSession, nil
}

func (c *fakeCeremony) FinishRegistration(domain.User, []domain.WebAuthnCredential, domain.WebAuthnCeremonySession, domain.WebAuthnCeremonyResponse) ([]byte, []byte, []byte, string, uint32, []string, bool, error) {
	return c.finishRegCredentialID, c.finishRegPublicKey, []byte{0}, c.finishRegAttestationType, 0, []string{"internal"}, false, nil
}

func (c *fakeCeremony) BeginLogin() (domain.WebAuthnCeremonyOptions, domain.WebAuthnCeremonySession, error) {
	return c.beginLoginOptions, c.beginLoginSession, nil
}

func (c *fakeCeremony) FinishLogin(handler webauthn_credential.DiscoverableUserHandler, _ domain.WebAuthnCeremonySession, _ domain.WebAuthnCeremonyResponse) (domain.User, []byte, uint32, error) {
	u, _, err := handler(c.loginRawID, nil)
	if err != nil {
		return domain.User{}, nil, 0, err
	}
	return u, c.finishLoginCredentialID, c.finishLoginSignCount, nil
}

// fakeChallengeTokens implements webauthn_credential.ChallengeTokens.
type fakeChallengeTokens struct {
	userID uuid.UUID
}

func (c *fakeChallengeTokens) IssueRegistration(uuid.UUID, json.RawMessage) (string, error) {
	return "reg-token", nil
}
func (c *fakeChallengeTokens) VerifyRegistration(string) (uuid.UUID, json.RawMessage, error) {
	return c.userID, json.RawMessage(`{"challenge":"abc"}`), nil
}
func (c *fakeChallengeTokens) IssueLogin(json.RawMessage) (string, error) {
	return "login-token", nil
}
func (c *fakeChallengeTokens) VerifyLogin(string) (json.RawMessage, error) {
	return json.RawMessage(`{"challenge":"xyz"}`), nil
}

func decodeData(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	var body struct {
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.NoError(t, json.Unmarshal(body.Data, v))
}

// newTestDeps wires a real kernel.Deps the same way cmd/server/deps.go does,
// but against local hand-written fakes instead of Postgres/go-webauthn — see
// routers/auth/tests' newTestDeps for the precedent this mirrors.
func newTestDeps(t *testing.T, u domain.User, credentials ...domain.WebAuthnCredential) (*kernel.Deps, string, *fakeCredentialRepo, *fakeCeremony) {
	t.Helper()
	hasher := security.NewArgon2PasswordHasher()
	hash, err := hasher.Hash("correct horse battery staple")
	require.NoError(t, err)
	u.PasswordHash = hash

	userRepo := &fakeUserRepo{byID: map[uuid.UUID]domain.User{u.ID: u}}
	credRepo := newFakeCredentialRepo(credentials...)
	uowManager := fakeUoWManager{}

	jwtConfig := jwtadapter.Config{SecretKey: "test-secret-key-not-for-production", Algorithm: "HS256", AccessTokenExpireMinutes: 60}
	tokenService := jwtadapter.NewAccessTokenService(jwtConfig)

	userService := user.NewService(userRepo, hasher, uowManager)
	credService := webauthn_credential.NewService(credRepo, uowManager)
	ceremony := &fakeCeremony{}
	challenges := &fakeChallengeTokens{userID: u.ID}

	deps := &kernel.Deps{
		AuthService:      auth.NewService(userRepo, hasher, tokenService),
		UserService:      userService,
		WebAuthnWorkflow: workflow.NewWebAuthnWorkflow(userService, credService, ceremony, challenges, tokenService),
	}

	token, err := tokenService.Issue(u.ID)
	require.NoError(t, err)
	return deps, token, credRepo, ceremony
}

func newRouter(deps *kernel.Deps) *chi.Mux {
	r := chi.NewRouter()
	protected := r.With(kernel.RequireAuth(deps))
	webauthnrouter.Mount(r, protected, deps)
	return r
}

func TestRegisterBeginHandler_HappyPath(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi"}
	deps, token, _, ceremony := newTestDeps(t, u)
	ceremony.beginRegOptions = []byte(`{"publicKey":{}}`)
	r := newRouter(deps)

	req := httptest.NewRequest(http.MethodPost, "/v1/webauthn/credentials/register/begin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp schema.WebAuthnRegisterBeginResponse
	decodeData(t, rec, &resp)
	assert.Equal(t, "reg-token", resp.ChallengeToken)
	assert.JSONEq(t, `{"publicKey":{}}`, string(resp.Options))
}

func TestRegisterBeginHandler_NoAuthIsUnauthorized(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi"}
	deps, _, _, _ := newTestDeps(t, u)
	r := newRouter(deps)

	req := httptest.NewRequest(http.MethodPost, "/v1/webauthn/credentials/register/begin", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRegisterFinishHandler_HappyPathPersistsCredential(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi"}
	deps, token, credRepo, ceremony := newTestDeps(t, u)
	ceremony.finishRegCredentialID = []byte("cred-1")
	ceremony.finishRegPublicKey = []byte("pub-1")
	ceremony.finishRegAttestationType = "none"
	r := newRouter(deps)

	body := `{"challenge_token":"reg-token","name":"iPhone","credential":{"id":"..."}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/webauthn/credentials/register/finish", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	var resp schema.WebAuthnCredentialResponse
	decodeData(t, rec, &resp)
	assert.Equal(t, "iPhone", resp.Name)
	assert.Len(t, credRepo.byID, 1)
}

func TestRegisterFinishHandler_BadJSON(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi"}
	deps, token, _, _ := newTestDeps(t, u)
	r := newRouter(deps)

	req := httptest.NewRequest(http.MethodPost, "/v1/webauthn/credentials/register/finish", bytes.NewBufferString(`{`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegisterFinishHandler_MissingName_400(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi"}
	deps, token, _, _ := newTestDeps(t, u)
	r := newRouter(deps)

	body := `{"challenge_token":"reg-token","name":"","credential":{"id":"..."}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/webauthn/credentials/register/finish", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListCredentialsHandler_OnlyListsCallersOwn(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi"}
	other := uuid.New()
	mine := domain.WebAuthnCredential{ID: uuid.New(), UserID: u.ID, Name: "iPhone", CredentialID: []byte("c1"), PublicKey: []byte("p1")}
	notMine := domain.WebAuthnCredential{ID: uuid.New(), UserID: other, Name: "Someone else's", CredentialID: []byte("c2"), PublicKey: []byte("p2")}
	deps, token, _, _ := newTestDeps(t, u, mine, notMine)
	r := newRouter(deps)

	req := httptest.NewRequest(http.MethodGet, "/v1/webauthn/credentials", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp []schema.WebAuthnCredentialResponse
	decodeData(t, rec, &resp)
	require.Len(t, resp, 1)
	assert.Equal(t, "iPhone", resp[0].Name)
}

func TestDeleteCredentialHandler_BelongsToAnotherUser_404(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi"}
	other := uuid.New()
	notMine := domain.WebAuthnCredential{ID: uuid.New(), UserID: other, Name: "Someone else's", CredentialID: []byte("c2"), PublicKey: []byte("p2")}
	deps, token, credRepo, _ := newTestDeps(t, u, notMine)
	r := newRouter(deps)

	req := httptest.NewRequest(http.MethodDelete, "/v1/webauthn/credentials/"+notMine.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	_, stillThere := credRepo.byID[notMine.ID]
	assert.True(t, stillThere)
}

func TestDeleteCredentialHandler_HappyPath(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi"}
	mine := domain.WebAuthnCredential{ID: uuid.New(), UserID: u.ID, Name: "iPhone", CredentialID: []byte("c1"), PublicKey: []byte("p1")}
	deps, token, credRepo, _ := newTestDeps(t, u, mine)
	r := newRouter(deps)

	req := httptest.NewRequest(http.MethodDelete, "/v1/webauthn/credentials/"+mine.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	_, stillThere := credRepo.byID[mine.ID]
	assert.False(t, stillThere)
}

func TestLoginBeginHandler_IsPublic(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi"}
	deps, _, _, ceremony := newTestDeps(t, u)
	ceremony.beginLoginOptions = []byte(`{"publicKey":{}}`)
	r := newRouter(deps)

	req := httptest.NewRequest(http.MethodPost, "/v1/webauthn/login/begin", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp schema.WebAuthnLoginBeginResponse
	decodeData(t, rec, &resp)
	assert.Equal(t, "login-token", resp.ChallengeToken)
}

func TestLoginFinishHandler_HappyPathIssuesAccessToken(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi"}
	cred := domain.WebAuthnCredential{ID: uuid.New(), UserID: u.ID, Name: "iPhone", CredentialID: []byte("cred-1"), PublicKey: []byte("pub-1")}
	deps, _, _, ceremony := newTestDeps(t, u, cred)
	ceremony.loginRawID = cred.CredentialID
	ceremony.finishLoginCredentialID = cred.CredentialID
	r := newRouter(deps)

	body := `{"challenge_token":"login-token","credential":{"id":"..."}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/webauthn/login/finish", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp schema.TokenResponse
	decodeData(t, rec, &resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, "bearer", resp.TokenType)
}

func TestLoginFinishHandler_UnknownCredential_Unauthorized(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi"}
	deps, _, _, ceremony := newTestDeps(t, u) // no credentials seeded
	ceremony.loginRawID = []byte("does-not-exist")
	r := newRouter(deps)

	body := `{"challenge_token":"login-token","credential":{"id":"..."}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/webauthn/login/finish", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLoginFinishHandler_BadJSON(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi"}
	deps, _, _, _ := newTestDeps(t, u)
	r := newRouter(deps)

	req := httptest.NewRequest(http.MethodPost, "/v1/webauthn/login/finish", bytes.NewBufferString(`{`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
