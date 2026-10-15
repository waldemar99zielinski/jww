package gateway

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

// PostgresTransactionCategoryRepository implements
// service/transactioncategory.Repository.
type PostgresTransactionCategoryRepository struct {
	db DBTX
}

func NewPostgresTransactionCategoryRepository(db DBTX) *PostgresTransactionCategoryRepository {
	return &PostgresTransactionCategoryRepository{db: db}
}

func scanTransactionCategoryRow(row pgx.Row) (*domain.TransactionCategory, error) {
	var c domain.TransactionCategory
	var categoryType string
	err := row.Scan(&c.ID, &categoryType, &c.Name, &c.Active, &c.SortOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.Type = domain.TransactionCategoryType(categoryType)
	return &c, nil
}

func (r *PostgresTransactionCategoryRepository) Create(ctx domain.AppContext, category domain.TransactionCategory, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Insert("transaction_categories").
		Columns("id", "type", "name", "active", "sort_order").
		Values(category.ID, string(category.Type), category.Name, category.Active, category.SortOrder).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}

func (r *PostgresTransactionCategoryRepository) GetByID(ctx domain.AppContext, categoryID uuid.UUID, uow ...unit_of_work.UnitOfWork) (*domain.TransactionCategory, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select("id", "type", "name", "active", "sort_order").
		From("transaction_categories").
		Where("id = ?", categoryID).
		ToSql()
	if err != nil {
		return nil, err
	}
	return scanTransactionCategoryRow(db.QueryRow(ctx.Context(), sql, args...))
}

func (r *PostgresTransactionCategoryRepository) ListByType(ctx domain.AppContext, categoryType domain.TransactionCategoryType, activeOnly bool, uow ...unit_of_work.UnitOfWork) ([]domain.TransactionCategory, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	builder := psql.Select("id", "type", "name", "active", "sort_order").
		From("transaction_categories").
		Where("type = ?", string(categoryType))
	if activeOnly {
		builder = builder.Where("active = true")
	}
	sql, args, err := builder.OrderBy("sort_order", "name").ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(ctx.Context(), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.TransactionCategory
	for rows.Next() {
		category, err := scanTransactionCategoryRow(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, *category)
	}
	return categories, rows.Err()
}

func (r *PostgresTransactionCategoryRepository) Update(ctx domain.AppContext, category domain.TransactionCategory, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Update("transaction_categories").
		Set("type", string(category.Type)).
		Set("name", category.Name).
		Set("active", category.Active).
		Set("sort_order", category.SortOrder).
		Where("id = ?", category.ID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}

func (r *PostgresTransactionCategoryRepository) Delete(ctx domain.AppContext, category domain.TransactionCategory, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Delete("transaction_categories").Where("id = ?", category.ID).ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}
