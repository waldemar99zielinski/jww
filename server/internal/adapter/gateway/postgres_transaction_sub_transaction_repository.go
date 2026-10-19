package gateway

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

// PostgresTransactionSubTransactionRepository implements
// transaction_sub_transaction.Repository.
type PostgresTransactionSubTransactionRepository struct {
	db DBTX
}

func NewPostgresTransactionSubTransactionRepository(db DBTX) *PostgresTransactionSubTransactionRepository {
	return &PostgresTransactionSubTransactionRepository{db: db}
}

const transactionSubTransactionSelectColumns = `
	id, transaction_id, name, label, amount, created_at`

func scanTransactionSubTransactionRow(row pgx.Row) (*domain.TransactionSubTransaction, error) {
	var s domain.TransactionSubTransaction
	var amountNumeric pgtype.Numeric

	err := row.Scan(
		&s.ID, &s.TransactionID, &s.Name, &s.Label, &amountNumeric, &s.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	amount, err := numericToMoney(amountNumeric)
	if err != nil {
		return nil, err
	}
	s.Amount = amount
	return &s, nil
}

func (r *PostgresTransactionSubTransactionRepository) Create(ctx domain.AppContext, sub domain.TransactionSubTransaction, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	amountNumeric, err := moneyToNumeric(sub.Amount)
	if err != nil {
		return err
	}
	sql, args, err := psql.Insert("transaction_sub_transactions").
		Columns("id", "transaction_id", "name", "label", "amount", "created_at").
		Values(sub.ID, sub.TransactionID, sub.Name, sub.Label, amountNumeric, sub.CreatedAt).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}

func (r *PostgresTransactionSubTransactionRepository) Update(ctx domain.AppContext, sub domain.TransactionSubTransaction, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	amountNumeric, err := moneyToNumeric(sub.Amount)
	if err != nil {
		return err
	}
	sql, args, err := psql.Update("transaction_sub_transactions").
		Set("name", sub.Name).
		Set("label", sub.Label).
		Set("amount", amountNumeric).
		Where("id = ?", sub.ID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}

func (r *PostgresTransactionSubTransactionRepository) Delete(ctx domain.AppContext, sub domain.TransactionSubTransaction, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Delete("transaction_sub_transactions").Where("id = ?", sub.ID).ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}

func (r *PostgresTransactionSubTransactionRepository) GetByID(ctx domain.AppContext, id uuid.UUID, uow ...unit_of_work.UnitOfWork) (*domain.TransactionSubTransaction, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select(transactionSubTransactionSelectColumns[1:]).
		From("transaction_sub_transactions").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return nil, err
	}
	return scanTransactionSubTransactionRow(db.QueryRow(ctx.Context(), sql, args...))
}

// ListByTransactionIDs returns every line item of the requested transactions in
// a stable (created_at, id) order. The ANY form binds the whole id list as one
// Postgres array parameter.
func (r *PostgresTransactionSubTransactionRepository) ListByTransactionIDs(ctx domain.AppContext, transactionIDs []int64, uow ...unit_of_work.UnitOfWork) ([]domain.TransactionSubTransaction, error) {
	if len(transactionIDs) == 0 {
		return nil, nil
	}
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select(transactionSubTransactionSelectColumns[1:]).
		From("transaction_sub_transactions").
		Where("transaction_id = ANY(?)", transactionIDs).
		OrderBy("created_at", "id").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(ctx.Context(), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []domain.TransactionSubTransaction
	for rows.Next() {
		s, err := scanTransactionSubTransactionRow(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, *s)
	}
	return subs, rows.Err()
}
