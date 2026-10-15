package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/adapter/security"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	authrouter "jww/internal/infrastructure/http_server/routers/auth"
	"jww/internal/service/auth"
	"jww/internal/service/unit_of_work"
)

// fakeRepository implements auth.UserRepository (a package-local, smaller
// interface than user.Repository — see auth.go's doc comment).
type fakeRepository struct {
	byUsername map[string]domain.User
}

func (f *fakeRepository) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.User, error) {
	for _, u := range f.byUsername {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, nil
}

func (f *fakeRepository) GetByUsername(_ domain.AppContext, username string, _ ...unit_of_work.UnitOfWork) (*domain.User, error) {
	u, ok := f.byUsername[username]
	if !ok {
		return nil, nil
	}
	return &u, nil
}

// This package uses the REAL security.Argon2PasswordHasher and
// security.JwtTokenService (not fakes) so get_me can be exercised behind the
// REAL kernel.RequireAuth middleware with a genuinely signed token — the
// only way to reach the "authenticated user is on the request context" path
// without reimplementing kernel's unexported context helpers.
func newTestDeps(t *testing.T) (*kernel.Deps, *fakeRepository, string) {
	t.Helper()
	hasher := security.NewArgon2PasswordHasher()
	hash, err := hasher.Hash("correct horse battery staple")
	require.NoError(t, err)

	u := domain.User{ID: uuid.New(), Username: "wozi", PasswordHash: hash, DisplayName: "Wozi"}
	repo := &fakeRepository{byUsername: map[string]domain.User{u.Username: u}}

	tokenService := security.NewJwtTokenService(security.Config{
		SecretKey: "test-secret-key-not-for-production", Algorithm: "HS256", AccessTokenExpireMinutes: 60,
	})

	deps := &kernel.Deps{
		AuthService: auth.NewService(repo, hasher, tokenService),
	}
	return deps, repo, u.ID.String()
}

func decodeData(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	var body struct {
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.NoError(t, json.Unmarshal(body.Data, v))
}

func TestLoginHandler(t *testing.T) {
	deps, _, _ := newTestDeps(t)
	r := chi.NewRouter()
	authrouter.Mount(r, r, deps) // no RequireAuth: login is public in the real Mount too

	t.Run("happy path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(`{"username":"wozi","password":"correct horse battery staple"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp schema.TokenResponse
		decodeData(t, rec, &resp)
		assert.NotEmpty(t, resp.AccessToken)
		assert.Equal(t, "bearer", resp.TokenType)
	})

	t.Run("wrong password is unauthorized, not a 500 or 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(`{"username":"wozi","password":"wrong"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("unknown username gets the same generic unauthorized (anti-enumeration)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(`{"username":"nobody","password":"whatever"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("bad JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(`{`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetMeHandler(t *testing.T) {
	deps, _, _ := newTestDeps(t)
	r := chi.NewRouter()
	protected := r.With(kernel.RequireAuth(deps))
	authrouter.Mount(r, protected, deps)

	t.Run("happy path with a real signed token", func(t *testing.T) {
		token, err := deps.AuthService.Login(domain.AppContext{}, "wozi", "correct horse battery staple")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp schema.UserResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, "wozi", resp.Username)
	})

	t.Run("missing Authorization header is unauthorized with the Bearer challenge", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"))
	})

	t.Run("garbage token is unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer not-a-real-token")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
