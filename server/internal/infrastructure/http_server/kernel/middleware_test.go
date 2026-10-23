package kernel

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/security"
	jwtadapter "jww/internal/adapter/security/jwt"
	"jww/internal/domain"
	"jww/internal/service/auth"
	"jww/internal/service/unit_of_work"
)

// --- RequestLogger ---------------------------------------------------------
//
// The log level RequestLogger picks is a pure function (levelForStatus,
// tested directly below) — asserting the actual emitted log record would
// require mutating the shared, process-global utils/logger base handler,
// which is fragile under parallel tests. These tests only check the
// HTTP-observable behavior: flow-id handling and pass-through.

func TestRequestLogger_EchoesFlowIDHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	RequestLogger(next).ServeHTTP(rec, req)

	got := rec.Header().Get(flowIDHeader)
	assert.NotEmpty(t, got)
	_, err := uuid.Parse(got)
	assert.NoError(t, err, "flow id should be a valid UUID")
}

func TestRequestLogger_PreservesIncomingFlowID(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(flowIDHeader, "existing-flow-id")
	rec := httptest.NewRecorder()

	RequestLogger(next).ServeHTTP(rec, req)

	assert.Equal(t, "existing-flow-id", rec.Header().Get(flowIDHeader))
}

func TestRequestLogger_PassesThroughStatusAndBody(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("hello"))
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	RequestLogger(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Equal(t, "hello", rec.Body.String())
}

func TestRequestLogger_SkipsHealthEndpoints(t *testing.T) {
	for _, path := range []string{"/api/v1/health", "/api/v1/readiness"} {
		t.Run(path, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			RequestLogger(next).ServeHTTP(rec, req)

			assert.True(t, called)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Empty(t, rec.Header().Get(flowIDHeader), "health probes bypass the logger, so no flow id is echoed")
		})
	}
}

func TestIsHealthRoute(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/api/v1/health", true},
		{"/api/v1/readiness", true},
		{"/api/v1/health/", false},
		{"/api/v1/transactions", false},
		{"/health", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			assert.Equal(t, tt.want, isHealthRoute(req))
		})
	}
}

func TestLevelForStatus(t *testing.T) {
	tests := []struct {
		status int
		want   slog.Level
	}{
		{http.StatusOK, slog.LevelInfo},
		{http.StatusCreated, slog.LevelInfo},
		{http.StatusBadRequest, slog.LevelWarn},
		{http.StatusNotFound, slog.LevelWarn},
		{http.StatusInternalServerError, slog.LevelError},
		{http.StatusBadGateway, slog.LevelError},
	}
	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.want, levelForStatus(tt.status))
		})
	}
}

// --- Recoverer ---------------------------------------------------------

func TestRecoverer_RecoversPanicAndWrites500(t *testing.T) {
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		Recoverer(panicking).ServeHTTP(rec, req)
	})

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var body ErrorEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, CodeInternal, body.Error.Code)
}

func TestRecoverer_PassesThroughWhenNoPanic(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	Recoverer(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// --- LimitRequestBody ---------------------------------------------------------

func TestLimitRequestBody(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := DecodeJSON[payload](r); err != nil {
			WriteBadRequest(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	t.Run("body under the limit succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()

		LimitRequestBody(10)(next).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("body over the limit fails as a bad request", func(t *testing.T) {
		body := strings.NewReader(`{"name":"this payload is well over ten bytes"}`)
		req := httptest.NewRequest(http.MethodPost, "/", body)
		rec := httptest.NewRecorder()

		LimitRequestBody(10)(next).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("non-positive maxBytes falls back to the default cap", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		rec := httptest.NewRecorder()

		LimitRequestBody(0)(next).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestLimitRequestBody_ExemptsAttachmentUploadRoutes(t *testing.T) {
	echo := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})
	big := strings.Repeat("x", 256)

	t.Run("standalone draft upload is exempt", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/transaction-attachments", strings.NewReader(big))
		rec := httptest.NewRecorder()
		LimitRequestBody(10)(echo).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Len(t, rec.Body.String(), len(big))
	})

	t.Run("nested upload is exempt", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/42/attachments", strings.NewReader(big))
		rec := httptest.NewRecorder()
		LimitRequestBody(10)(echo).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Len(t, rec.Body.String(), len(big))
	})

	t.Run("ordinary endpoints stay capped", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", strings.NewReader(big))
		rec := httptest.NewRecorder()
		LimitRequestBody(10)(echo).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("multipart is not exempt globally", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/some/other", strings.NewReader(big))
		req.Header.Set("Content-Type", "multipart/form-data; boundary=x")
		rec := httptest.NewRecorder()
		LimitRequestBody(10)(echo).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestIsAttachmentUploadRoute(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   bool
	}{
		{http.MethodPost, "/api/v1/transaction-attachments", true},
		{http.MethodPost, "/api/v1/transactions/42/attachments", true},
		{http.MethodGet, "/api/v1/transaction-attachments", false},
		{http.MethodPost, "/api/v1/transactions", false},
		{http.MethodPost, "/api/v1/transactions/42/attachments/7", false},
		{http.MethodPost, "/api/v1/transactions/42/attachments/7/download", false},
		{http.MethodPut, "/api/v1/transactions/42/attachments/7", false},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			assert.Equal(t, tt.want, isAttachmentUploadRoute(req))
		})
	}
}

// --- bearerToken / ensureFlowID / RequestContext ---------------------------------------------------------

func TestBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantToken string
		wantOK    bool
	}{
		{"valid bearer", "Bearer abc123", "abc123", true},
		{"missing header", "", "", false},
		{"wrong scheme", "Basic abc123", "", false},
		{"empty token", "Bearer ", "", false},
		{"case-insensitive scheme", "bearer abc123", "abc123", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			token, ok := bearerToken(req)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantToken, token)
		})
	}
}

func TestEnsureFlowID(t *testing.T) {
	t.Run("generates one when absent and stores it on the request header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		got := ensureFlowID(req)

		assert.NotEmpty(t, got)
		assert.Equal(t, got, req.Header.Get(flowIDHeader))
	})

	t.Run("reuses an incoming header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(flowIDHeader, "incoming-id")

		assert.Equal(t, "incoming-id", ensureFlowID(req))
	})

	t.Run("prefers a value already on the context over the header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(flowIDHeader, "header-id")
		req = req.WithContext(withFlowID(req.Context(), "context-id"))

		assert.Equal(t, "context-id", ensureFlowID(req))
	})
}

func TestRequestContext(t *testing.T) {
	t.Run("unauthenticated request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/currencies", nil)

		appCtx := RequestContext(req)

		assert.Nil(t, appCtx.User())
		assert.Equal(t, domain.RequestSourceTypeHTTP, appCtx.RequestSource().Type)
		require.NotNil(t, appCtx.RequestSource().HTTP)
		assert.Equal(t, "/v1/currencies", appCtx.RequestSource().HTTP.Endpoint)
		assert.NotEmpty(t, appCtx.FlowID())
	})

	t.Run("authenticated request carries the user", func(t *testing.T) {
		user := &domain.User{ID: uuid.New(), Username: "alice"}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(withUser(req.Context(), user))

		appCtx := RequestContext(req)

		require.NotNil(t, appCtx.User())
		assert.Equal(t, user.ID, appCtx.User().ID)
	})
}

// --- RequireAuth ---------------------------------------------------------
//
// auth.Service is built from its own package-local ports (UserRepository,
// PasswordHasher, TokenService) — same situation as workflow (see
// TESTING.md): it holds concrete collaborators, not swappable interfaces at
// this level, so the test wires a *real* auth.Service. The repository is
// faked by hand (auth.UserRepository, deliberately narrower than
// service/user.Repository — see internal/service/auth/auth.go), but the
// PasswordHasher and TokenService are the real security adapters
// (security.Argon2PasswordHasher, jwt.AccessTokenService) with a
// throwaway secret — no real I/O, pure in-memory crypto, so this stays a
// fast unit test while still proving RequireAuth against genuine JWT
// signature/expiry checks rather than a hand-rolled stand-in for them.

type fakeUserRepository struct {
	usersByID map[uuid.UUID]*domain.User
}

func (f *fakeUserRepository) GetByID(_ domain.AppContext, userID uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.User, error) {
	return f.usersByID[userID], nil
}

func (f *fakeUserRepository) GetByUsername(_ domain.AppContext, _ string, _ ...unit_of_work.UnitOfWork) (*domain.User, error) {
	return nil, nil
}

func TestRequireAuth(t *testing.T) {
	const throwawaySecret = "test-only-throwaway-secret"

	userID := uuid.New()
	storedUser := &domain.User{ID: userID, Username: "alice", PasswordHash: "hash", DisplayName: "Alice"}
	repo := &fakeUserRepository{usersByID: map[uuid.UUID]*domain.User{userID: storedUser}}
	tokenSvc := jwtadapter.NewAccessTokenService(jwtadapter.Config{
		SecretKey:                throwawaySecret,
		Algorithm:                "HS256",
		AccessTokenExpireMinutes: 60,
	})
	deps := &Deps{AuthService: auth.NewService(repo, security.NewArgon2PasswordHasher(), tokenSvc)}

	downstream := func(t *testing.T, called *bool) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*called = true
			u := RequestContext(r).User()
			require.NotNil(t, u)
			assert.Equal(t, userID, u.ID)
			w.WriteHeader(http.StatusOK)
		})
	}

	t.Run("missing header rejects without calling downstream handler", func(t *testing.T) {
		var called bool
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		RequireAuth(deps)(downstream(t, &called)).ServeHTTP(rec, req)

		assert.False(t, called)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"))
	})

	t.Run("malformed header rejects", func(t *testing.T) {
		var called bool
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Basic xxx")
		rec := httptest.NewRecorder()

		RequireAuth(deps)(downstream(t, &called)).ServeHTTP(rec, req)

		assert.False(t, called)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("valid token for a user the repo doesn't have is rejected", func(t *testing.T) {
		unknownUserID := uuid.New()
		token, err := tokenSvc.Issue(unknownUserID)
		require.NoError(t, err)
		var called bool
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		RequireAuth(deps)(downstream(t, &called)).ServeHTTP(rec, req)

		assert.False(t, called)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"))
	})

	t.Run("token with an invalid signature is rejected", func(t *testing.T) {
		wrongSecretSvc := jwtadapter.NewAccessTokenService(jwtadapter.Config{
			SecretKey:                "a-different-secret",
			Algorithm:                "HS256",
			AccessTokenExpireMinutes: 60,
		})
		token, err := wrongSecretSvc.Issue(userID)
		require.NoError(t, err)
		var called bool
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		RequireAuth(deps)(downstream(t, &called)).ServeHTTP(rec, req)

		assert.False(t, called)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("valid token attaches the user to the request context", func(t *testing.T) {
		token, err := tokenSvc.Issue(userID)
		require.NoError(t, err)
		var called bool
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		RequireAuth(deps)(downstream(t, &called)).ServeHTTP(rec, req)

		assert.True(t, called)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
