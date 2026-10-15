// Package http_server is the HTTP composition root. It builds the chi router,
// attaches the cross-cutting middleware, and hands each resource subpackage
// both a public router and a protected router (already guarded by
// RequireAuth) so the subpackage decides where each of its routes belongs.
package http_server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"jww/internal/infrastructure/http_server/kernel"
	"jww/internal/infrastructure/http_server/routers/auth"
	"jww/internal/infrastructure/http_server/routers/currency"
	"jww/internal/infrastructure/http_server/routers/currency_exchange_rate"
	"jww/internal/infrastructure/http_server/routers/summary"
	"jww/internal/infrastructure/http_server/routers/transaction"
	"jww/internal/infrastructure/http_server/routers/transaction_category"
	"jww/internal/infrastructure/http_server/routers/user"
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
func New(deps *kernel.Deps, cfg Config) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr(),
		Handler:           NewRouter(deps),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
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
		// go-chi/cors expands "*" for origins and headers, but NOT for methods:
		// a literal ["*"] matches no real method, so isMethodAllowed fails and
		// the middleware returns before setting Access-Control-Allow-Origin.
		// List every method the API serves explicitly.
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
		// public and protected share the same routing tree; protected just
		// has RequireAuth prepended, so a resource routes each endpoint onto
		// whichever one it belongs to.
		public := api
		protected := api.With(kernel.RequireAuth(deps))

		auth.Mount(public, protected, deps)
		user.Mount(public, protected, deps)
		currency.Mount(public, protected, deps)
		currency_exchange_rate.Mount(public, protected, deps)
		transaction_category.Mount(public, protected, deps)
		transaction.Mount(public, protected, deps)
		summary.Mount(public, protected, deps)
	})

	return r
}
