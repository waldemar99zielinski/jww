package database

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

// PostgresUnitOfWorkManager creates Postgres-backed units of work.
type PostgresUnitOfWorkManager struct {
	pool *pgxpool.Pool
}

func NewPostgresUnitOfWorkManager(pool *pgxpool.Pool) *PostgresUnitOfWorkManager {
	return &PostgresUnitOfWorkManager{pool: pool}
}

func (m *PostgresUnitOfWorkManager) Create(ctx domain.AppContext) (unit_of_work.UnitOfWork, error) {
	tx, err := m.pool.Begin(ctx.Context())
	if err != nil {
		return nil, err
	}
	return &PostgresUnitOfWork{tx: tx}, nil
}

// PostgresUnitOfWork implements unit_of_work.UnitOfWork on top of a pgx
// transaction.
type PostgresUnitOfWork struct {
	tx pgx.Tx
}

func (u *PostgresUnitOfWork) Finish(ctx domain.AppContext) error {
	if u.tx == nil {
		return fmt.Errorf("database: Finish called on a unit of work that was never Started")
	}
	return u.tx.Commit(ctx.Context())
}

func (u *PostgresUnitOfWork) Restore(ctx domain.AppContext) error {
	if u.tx == nil {
		return fmt.Errorf("database: Restore called on a unit of work that was never Started")
	}
	return u.tx.Rollback(ctx.Context())
}

// GetValue returns the active pgx.Tx. Only the concrete Postgres gateways know
// how to interpret it.
func (u *PostgresUnitOfWork) GetValue() interface{} {
	return u.tx
}
