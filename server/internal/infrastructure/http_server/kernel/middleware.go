package kernel

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"jww/internal/domain"
)

// RequestLogger logs every request through the shared logger, binding the
// request's flow ID and source info from the AppContext. It resolves (or
// generates) the flow ID once, stores it on the request context, echoes it
// back in the response header, and then reuses that same ID when building
// the log entry — so middleware and handler logs share one flow. The log
// level follows the response status: 5xx → error, 4xx → warn, else info.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flowID := ensureFlowID(r)
		r = r.WithContext(withFlowID(r.Context(), flowID))
		w.Header().Set(flowIDHeader, flowID)

		wrapped := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		next.ServeHTTP(wrapped, r)

		appCtx := RequestContext(r)
		status := wrapped.Status()
		httpLog.WithContext(appCtx).LogAttrs(
			r.Context(),
			levelForStatus(status),
			"request completed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", status),
			slog.Int("response_bytes", wrapped.BytesWritten()),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

// Recoverer replaces chi's default Recoverer so recovered panics are emitted
// through the shared logger (with the request's flow ID) instead of straight
// to stderr, and the client still receives a 500.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				appCtx := RequestContext(r)
				httpLog.WithContext(appCtx).LogAttrs(
					r.Context(),
					slog.LevelError,
					"panic recovered",
					slog.Any("panic", rec),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("stack", string(debug.Stack())),
				)
				writeError(w, http.StatusInternalServerError, CodeInternal, "internal server error", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// defaultMaxBodyBytes caps a request body when Deps.MaxBodyBytes isn't set
// (e.g. a test building kernel.Deps{} directly), so every request path is
// protected even without explicit configuration.
const defaultMaxBodyBytes = 1 << 20 // 1 MiB

// LimitRequestBody caps every request body at maxBytes (or
// defaultMaxBodyBytes when maxBytes <= 0), so an oversized/slow body can't
// exhaust memory or hold a connection open indefinitely. A body that exceeds
// the limit fails with "http: request body too large" on the next Read,
// which DecodeJSON surfaces as a decode error → 400 via WriteBadRequest.
func LimitRequestBody(maxBytes int64) func(http.Handler) http.Handler {
	if maxBytes <= 0 {
		maxBytes = defaultMaxBodyBytes
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func levelForStatus(status int) slog.Level {
	switch {
	case status >= http.StatusInternalServerError:
		return slog.LevelError
	case status >= http.StatusBadRequest:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

// bearerToken extracts the token from "Authorization: Bearer <token>",
// reporting ok=false when the header is missing or malformed.
func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", false
	}
	return token, true
}

// RequireAuth rejects requests that don't carry a valid bearer token and, on
// success, stores the authenticated user on the request context so every
// downstream RequestContext (and its logs) observes it.
func RequireAuth(deps *Deps) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				WriteError(w, domain.NewAuthenticationError())
				return
			}
			user, err := deps.AuthService.Authenticate(RequestContext(r), token)
			if err != nil {
				WriteError(w, err)
				return
			}
			next.ServeHTTP(w, r.WithContext(withUser(r.Context(), user)))
		})
	}
}
