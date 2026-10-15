// Package addcurrency creates currencies directly in the database, bypassing
// the HTTP API, using the same gateway repository and currency service the
// server uses.
//
// It holds the reusable half of the add-currency script: the command that
// calls it lives in cmd/, and scripts/bootstrap reuses Client to seed the
// local development currencies.
package addcurrency

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"jww/internal/adapter/gateway"
	"jww/internal/domain"
	"jww/internal/infrastructure/database"
	"jww/internal/service/currency"
)

// Client reads and writes currencies through the server's service stack.
type Client struct {
	currencies *currency.Service
	repo       *gateway.PostgresCurrencyRepository
}

// NewClient wires the currency service over pool using the same adapters as
// the server. The concrete repository is kept alongside the service because
// the tools also need a lookup (GetByCode) that returns nil rather than an
// error when the currency does not exist.
func NewClient(pool *pgxpool.Pool) *Client {
	repo := gateway.NewPostgresCurrencyRepository(pool)
	return &Client{
		currencies: currency.NewService(repo, database.NewPostgresUnitOfWorkManager(pool)),
		repo:       repo,
	}
}

// GetByCode returns the stored currency, or nil when none exists.
func (c *Client) GetByCode(ctx domain.AppContext, code string) (*domain.Currency, error) {
	return c.repo.GetByCode(ctx, code)
}

// Create inserts a single currency with the given name and base flag.
func (c *Client) Create(ctx domain.AppContext, code, name string, isBase bool) (*domain.Currency, error) {
	return c.currencies.Create(ctx, code, name, isBase)
}
