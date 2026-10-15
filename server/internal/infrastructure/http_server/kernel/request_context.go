package kernel

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"jww/internal/domain"
)

// headerPrefix is prepended to every custom application header.
const headerPrefix = "X-JWW"

const flowIDHeader = headerPrefix + "-Flow-ID"

type flowIDKey struct{}

type userKey struct{}

// withUser stores the authenticated user on the request context so
// RequestContext exposes it to handlers and the logger.
func withUser(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

func userFromContext(ctx context.Context) *domain.User {
	user, _ := ctx.Value(userKey{}).(*domain.User)
	return user
}

// newFlowID returns a new RFC 4122 UUIDv4, used to tag a request that carried
// no flow ID of its own.
func newFlowID() string {
	return uuid.NewString()
}

// ensureFlowID resolves a request's flow ID, in priority order: the value
// already stored on the context (set by the logging middleware), the incoming
// X-JWW-Flow-ID header, or a freshly generated UUIDv4. When a new ID is
// generated it is added to the request headers so the rest of the request
// observes the same value.
func ensureFlowID(r *http.Request) string {
	flowID := flowIDFromContext(r.Context())
	if flowID == "" {
		flowID = r.Header.Get(flowIDHeader)
	}
	if flowID == "" {
		flowID = newFlowID()
		r.Header.Set(flowIDHeader, flowID)
	}
	return flowID
}

// withFlowID stores the request's flow ID so both the logging middleware and
// RequestContext observe the same one.
func withFlowID(ctx context.Context, flowID string) context.Context {
	return context.WithValue(ctx, flowIDKey{}, flowID)
}

func flowIDFromContext(ctx context.Context) string {
	if flowID, ok := ctx.Value(flowIDKey{}).(string); ok {
		return flowID
	}
	return ""
}

// RequestContext builds the domain.AppContext for one request, reusing the
// flow ID already resolved by the request logging middleware when present.
func RequestContext(r *http.Request) domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(r.Context())
	appCtx.SetRequestSource(domain.RequestSource{
		Type: domain.RequestSourceTypeHTTP,
		HTTP: &domain.HTTPRequestSource{Endpoint: r.URL.Path},
	})
	appCtx.SetUser(userFromContext(r.Context()))
	appCtx.SetFlowID(ensureFlowID(r))
	return appCtx
}
