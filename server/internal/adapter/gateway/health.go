package gateway

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"jww/internal/domain"
)

// PostgresHealthChecker adapts the Postgres connection pool to the health
// service's Checker port.
type PostgresHealthChecker struct {
	pool *pgxpool.Pool
}

func NewPostgresHealthChecker(pool *pgxpool.Pool) *PostgresHealthChecker {
	return &PostgresHealthChecker{pool: pool}
}

// Check verifies that PostgreSQL can currently accept a connection.
func (c *PostgresHealthChecker) Check(ctx domain.AppContext) error {
	if err := c.pool.Ping(ctx.Context()); err != nil {
		return fmt.Errorf("gateway: database health check: %w", err)
	}
	return nil
}
