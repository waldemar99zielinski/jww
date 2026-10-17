//go:build e2e

// Package e2e_test holds the server's full-stack HTTP journey tests — real
// Postgres (via internal/tests/pgtest), real adapters, no fakes anywhere.
// See TESTING.md's "End-to-end" section and this directory's README.md.
//
// This directory contains only test files; there is no production "e2e"
// package, so every file here uses the external test package name
// (package e2e_test) uniformly.
package e2e_test

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"jww/internal/domain"
	http_server "jww/internal/infrastructure/http_server"
	"jww/internal/tests/pgtest"
	"jww/internal/tests/s3test"
	"jww/tests/utils"
)

// testJWTSecret is a throwaway signing key for the duration of one test
// binary run — never a real credential.
const testJWTSecret = "e2e-test-secret-key-do-not-use-elsewhere"

// seededPassword is the password for the one user account seeded in
// TestMain and reused by every journey that needs an authenticated client.
const seededPassword = "e2e-seeded-user-password"

var (
	pgPool     *pgtest.Pool
	minioStore *s3test.Bucket
	httpSrv    *httptest.Server
	apiClient  *utils.Client
	seededUser *domain.User
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pool, err := pgtest.Start(ctx)
	if err != nil {
		panic("e2e: starting Postgres: " + err.Error())
	}
	pgPool = pool

	store, err := s3test.Start(ctx, "jww-e2e")
	if err != nil {
		pool.Close(ctx)
		panic("e2e: starting MinIO: " + err.Error())
	}
	minioStore = store

	deps, err := utils.BuildTestDeps(ctx, pool.Pool, testJWTSecret, store.Config)
	if err != nil {
		store.Close(ctx)
		pool.Close(ctx)
		panic("e2e: building deps: " + err.Error())
	}

	appCtx := pgtest.AppContext(ctx)
	created, err := deps.UserService.Create(appCtx, "e2e-seeded-user", seededPassword, "E2E Seeded User")
	if err != nil {
		store.Close(ctx)
		pool.Close(ctx)
		panic("e2e: seeding user: " + err.Error())
	}
	seededUser = created

	httpSrv = httptest.NewServer(http_server.NewRouter(deps))
	apiClient = utils.NewClient(httpSrv.URL, nil)

	code := m.Run()

	httpSrv.Close()
	store.Close(ctx)
	pool.Close(ctx)
	os.Exit(code)
}
