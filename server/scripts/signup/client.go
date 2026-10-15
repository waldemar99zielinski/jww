// Package signup creates users directly in the database, bypassing the HTTP
// API, using the same gateway repository, Argon2id hasher, and user service the
// server uses — so the rows it inserts log in through /v1/auth/login like any
// other.
//
// It holds the reusable half of the signup script: the command that calls it
// lives in cmd/, and scripts/bootstrap reuses Client to seed the local test
// accounts.
package signup

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"jww/internal/adapter/gateway"
	"jww/internal/adapter/security"
	"jww/internal/domain"
	"jww/internal/infrastructure/database"
	"jww/internal/service/user"
)

// Client reads and writes users through the server's service stack.
type Client struct {
	users *user.Service
	repo  *gateway.PostgresUserRepository
}

// NewClient wires the user service over pool using the same adapters as the
// server. The concrete repository is kept alongside the service because the
// tools also need a lookup (GetByUsername) that the user service's port does
// not expose.
func NewClient(pool *pgxpool.Pool) *Client {
	repo := gateway.NewPostgresUserRepository(pool)
	return &Client{
		users: user.NewService(
			repo,
			security.NewArgon2PasswordHasher(),
			database.NewPostgresUnitOfWorkManager(pool),
		),
		repo: repo,
	}
}

// GetByUsername returns the stored account, or nil when none exists.
func (c *Client) GetByUsername(ctx domain.AppContext, username string) (*domain.User, error) {
	return c.repo.GetByUsername(ctx, username)
}

// Create hashes password and inserts a single account with the given display
// name.
func (c *Client) Create(ctx domain.AppContext, username, password, displayName string) (*domain.User, error) {
	return c.users.Create(ctx, username, password, displayName)
}
