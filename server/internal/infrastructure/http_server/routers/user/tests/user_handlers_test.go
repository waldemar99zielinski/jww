package user_test

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
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	userrouter "jww/internal/infrastructure/http_server/routers/user"
	"jww/internal/service/unit_of_work"
	"jww/internal/service/user"
)

type fakeRepository struct {
	byID map[uuid.UUID]domain.User
}

func newFakeRepository() *fakeRepository { return &fakeRepository{byID: map[uuid.UUID]domain.User{}} }

func (f *fakeRepository) Create(_ domain.AppContext, u domain.User, _ ...unit_of_work.UnitOfWork) error {
	f.byID[u.ID] = u
	return nil
}

func (f *fakeRepository) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, nil
	}
	return &u, nil
}

func (f *fakeRepository) List(_ domain.AppContext, _ ...unit_of_work.UnitOfWork) ([]domain.User, error) {
	out := make([]domain.User, 0, len(f.byID))
	for _, u := range f.byID {
		out = append(out, u)
	}
	return out, nil
}

func (f *fakeRepository) Update(_ domain.AppContext, u domain.User, _ ...unit_of_work.UnitOfWork) error {
	f.byID[u.ID] = u
	return nil
}

func (f *fakeRepository) Delete(_ domain.AppContext, u domain.User, _ ...unit_of_work.UnitOfWork) error {
	delete(f.byID, u.ID)
	return nil
}

type fakePasswordHasher struct{}

func (fakePasswordHasher) Hash(password string) (string, error) { return "hashed:" + password, nil }
func (fakePasswordHasher) Verify(password, hash string) bool    { return "hashed:"+password == hash }

type fakeUoWManager struct{}

func (fakeUoWManager) Create(domain.AppContext) (unit_of_work.UnitOfWork, error) {
	return fakeUoW{}, nil
}

type fakeUoW struct{}

func (fakeUoW) Finish(domain.AppContext) error  { return nil }
func (fakeUoW) Restore(domain.AppContext) error { return nil }
func (fakeUoW) GetValue() interface{}           { return nil }

func newTestRouter(t *testing.T) (*chi.Mux, *fakeRepository) {
	t.Helper()
	repo := newFakeRepository()
	deps := &kernel.Deps{
		UserService: user.NewService(repo, fakePasswordHasher{}, fakeUoWManager{}),
	}
	r := chi.NewRouter()
	userrouter.Mount(r, r, deps)
	return r, repo
}

func decodeData(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	var body struct {
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.NoError(t, json.Unmarshal(body.Data, v))
}

func seedUser(repo *fakeRepository) domain.User {
	u := domain.User{ID: uuid.New(), Username: "wozi", PasswordHash: "hashed:secret", DisplayName: "Wozi"}
	_ = repo.Create(domain.AppContext{}, u)
	return u
}

func TestListUsersHandler(t *testing.T) {
	r, repo := newTestRouter(t)
	seedUser(repo)
	seedUser(repo)

	req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp []schema.UserResponse
	decodeData(t, rec, &resp)
	assert.Len(t, resp, 2)
}

func TestGetUserHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, repo := newTestRouter(t)
		u := seedUser(repo)

		req := httptest.NewRequest(http.MethodGet, "/v1/users/"+u.ID.String(), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp schema.UserResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, u.Username, resp.Username)
		assert.NotContains(t, rec.Body.String(), "password", "response must never leak the password hash")
	})

	t.Run("bad uuid", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodGet, "/v1/users/not-a-uuid", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodGet, "/v1/users/"+uuid.NewString(), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestUpdateUserHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, repo := newTestRouter(t)
		u := seedUser(repo)

		req := httptest.NewRequest(http.MethodPut, "/v1/users/"+u.ID.String(), bytes.NewBufferString(`{"display_name":"Woziu"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp schema.UserResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, "Woziu", resp.DisplayName)
	})

	t.Run("failed Validate", func(t *testing.T) {
		r, repo := newTestRouter(t)
		u := seedUser(repo)
		req := httptest.NewRequest(http.MethodPut, "/v1/users/"+u.ID.String(), bytes.NewBufferString(`{"display_name":""}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodPut, "/v1/users/"+uuid.NewString(), bytes.NewBufferString(`{"display_name":"X"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestDeleteUserHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, repo := newTestRouter(t)
		u := seedUser(repo)
		req := httptest.NewRequest(http.MethodDelete, "/v1/users/"+u.ID.String(), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodDelete, "/v1/users/"+uuid.NewString(), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
