// Package http_server is the HTTP composition root. It builds the chi router,
// attaches the cross-cutting middleware, and hands each resource subpackage
// both a public router and a protected router (already guarded by
// RequireAuth) so the subpackage decides where each of its routes belongs.
package http_server

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"jww/internal/infrastructure/http_server/kernel"
	"jww/internal/infrastructure/http_server/routers/auth"
	"jww/internal/infrastructure/http_server/routers/currency"
	"jww/internal/infrastructure/http_server/routers/currency_exchange_rate"
	"jww/internal/infrastructure/http_server/routers/health"
	"jww/internal/infrastructure/http_server/routers/summary"
	"jww/internal/infrastructure/http_server/routers/transaction"
	transaction_attachment "jww/internal/infrastructure/http_server/routers/transaction_attachment"
	"jww/internal/infrastructure/http_server/routers/transaction_category"
	"jww/internal/infrastructure/http_server/routers/user"
	"jww/internal/infrastructure/http_server/routers/webauthn"
	httpweb "jww/internal/infrastructure/http_server/web"
)

// Timeouts applied to every connection the server accepts. ReadHeaderTimeout
// in particular guards against Slowloris-style attacks (a client that opens a
// connection and trickles headers in to keep it alive); the others bound how
// long one request/response or an idle keep-alive connection may run.
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second
)

// New builds the HTTP server with every route mounted and ready to listen.
// It validates optional web serving here as well as during environment loading
// so direct callers cannot accidentally construct a partially working server.
func New(deps *kernel.Deps, cfg Config, webCfg httpweb.Config) (*http.Server, error) {
	handler, err := NewHandler(deps, webCfg)
	if err != nil {
		return nil, err
	}
	return &http.Server{
		Addr:              cfg.Addr(),
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}, nil
}

// NewHandler builds the complete HTTP application. A disabled web config
// returns the API-only router used by tests and local backend development.
func NewHandler(deps *kernel.Deps, webCfg httpweb.Config) (http.Handler, error) {
	var webHandler http.Handler
	if webCfg.Enabled() {
		var err error
		webHandler, err = httpweb.New(webCfg)
		if err != nil {
			return nil, err
		}
	} else {
		if err := webCfg.Validate(); err != nil {
			return nil, err
		}
	}

	api := NewRouter(deps)
	handler := api
	if webHandler != nil {
		handler = newApplicationHandler(api, webHandler)
	}
	return handler, nil
}

// newApplicationHandler is the top-level route boundary. Keeping this choice
// here means the web handler knows nothing about API routing and the API router
// knows nothing about SPA fallback behavior.
func newApplicationHandler(api, web http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
			api.ServeHTTP(w, r)
			return
		}
		web.ServeHTTP(w, r)
	})
}

// NewRouter wires every resource subpackage into one chi.Router. Each
// subpackage receives the public router and the protected router and mounts
// its routes on whichever fits; the middleware here never has to know which
// routes are protected.
func NewRouter(deps *kernel.Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(kernel.RequestLogger)
	r.Use(kernel.Recoverer)
	r.Use(kernel.LimitRequestBody(deps.MaxBodyBytes))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: deps.CorsOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{"*"},
	}))

	r.Route("/api", func(api chi.Router) {
		public := api
		protected := api.With(kernel.RequireAuth(deps))

		health.Mount(public, protected, deps)
		auth.Mount(public, protected, deps)
		user.Mount(public, protected, deps)
		currency.Mount(public, protected, deps)
		currency_exchange_rate.Mount(public, protected, deps)
		transaction_category.Mount(public, protected, deps)
		transaction.Mount(public, protected, deps)
		transaction_attachment.Mount(public, protected, deps)
		summary.Mount(public, protected, deps)
		webauthn.Mount(public, protected, deps)
	})

	return r
}
