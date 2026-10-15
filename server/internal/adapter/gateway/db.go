package gateway

import (
	"context"
	"fmt"
	"jww/internal/service/unit_of_work"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX is the subset of pgx.Tx and *pgxpool.Pool every gateway needs.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// psql is the Squirrel statement builder configured for Postgres's $N
// placeholder style, shared by every gateway in this package.
var psql = newStatementBuilder()

func dbFromUnitOfWork(defaultDB DBTX, uow ...unit_of_work.UnitOfWork) (DBTX, error) {
	if len(uow) == 0 || uow[0] == nil {
		return defaultDB, nil
	}
	if len(uow) > 1 {
		return nil, fmt.Errorf("gateway: expected at most one unit of work, got %d", len(uow))
	}
	db, ok := uow[0].GetValue().(DBTX)
	if !ok {
		return nil, fmt.Errorf("gateway: unexpected unit of work value %T", uow[0].GetValue())
	}
	return db, nil
}
