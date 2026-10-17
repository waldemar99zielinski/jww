package gateway

import (
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"jww/internal/domain"
	"jww/internal/service/transaction"
	"jww/internal/service/unit_of_work"
)

// PostgresTransactionRepository implements service/transaction.Repository.
type PostgresTransactionRepository struct {
	db DBTX
}

func NewPostgresTransactionRepository(db DBTX) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

const transactionSelectColumns = `
	t.id, t.type, t.amount, t.source, t.name, t.description, t.created_at,
	c.code, c.name, c.is_base,
	tc.id, tc.type, tc.name, tc.active, tc.sort_order,
	u.id, u.username, u.password_hash, u.display_name, u.created_at`

const transactionFrom = `
	transactions t
	JOIN currencies c ON c.code = t.currency_code
	JOIN transaction_categories tc ON tc.id = t.category_id
	JOIN users u ON u.id = t.user_id`

func scanTransactionRow(row pgx.Row) (*domain.Transaction, error) {
	var tr domain.Transaction
	var transactionType, source, categoryType string
	var amountNumeric pgtype.Numeric

	err := row.Scan(
		&tr.ID, &transactionType, &amountNumeric, &source, &tr.Name, &tr.Description, &tr.CreatedAt,
		&tr.Currency.Code, &tr.Currency.Name, &tr.Currency.IsBase,
		&tr.Category.ID, &categoryType, &tr.Category.Name, &tr.Category.Active, &tr.Category.SortOrder,
		&tr.User.ID, &tr.User.Username, &tr.User.PasswordHash, &tr.User.DisplayName, &tr.User.CreatedAt,
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
	tr.Type = domain.TransactionType(transactionType)
	tr.Source = domain.TransactionSource(source)
	tr.Category.Type = domain.TransactionCategoryType(categoryType)
	tr.Amount = amount
	return &tr, nil
}

func (r *PostgresTransactionRepository) Create(
	ctx domain.AppContext,
	transactionType domain.TransactionType,
	amount domain.Money,
	currency domain.Currency,
	category domain.TransactionCategory,
	user domain.User,
	source domain.TransactionSource,
	createdAt time.Time,
	name string,
	description string,
	uow ...unit_of_work.UnitOfWork,
) (*domain.Transaction, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	amountNumeric, err := moneyToNumeric(amount)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Insert("transactions").
		Columns("type", "amount", "currency_code", "category_id", "user_id", "source", "name", "description", "created_at").
		Values(string(transactionType), amountNumeric, currency.Code, category.ID, user.ID, string(source), name, description, createdAt).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return nil, err
	}
	var id int64
	if err := db.QueryRow(ctx.Context(), sql, args...).Scan(&id); err != nil {
		return nil, err
	}
	return &domain.Transaction{
		ID:          id,
		Type:        transactionType,
		Amount:      amount,
		Currency:    currency,
		Category:    category,
		User:        user,
		Source:      source,
		Name:        name,
		CreatedAt:   createdAt,
		Description: description,
	}, nil
}

func (r *PostgresTransactionRepository) GetByID(ctx domain.AppContext, transactionID int64, uow ...unit_of_work.UnitOfWork) (*domain.Transaction, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select(transactionSelectColumns[1:]).
		From(transactionFrom[1:]).
		Where("t.id = ?", transactionID).
		ToSql()
	if err != nil {
		return nil, err
	}
	return scanTransactionRow(db.QueryRow(ctx.Context(), sql, args...))
}

func (r *PostgresTransactionRepository) List(ctx domain.AppContext, filter *transaction.Filter, uow ...unit_of_work.UnitOfWork) ([]domain.Transaction, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	builder := psql.Select(transactionSelectColumns[1:]).From(transactionFrom[1:])

	if filter != nil {
		if filter.UserID != nil {
			builder = builder.Where("t.user_id = ?", *filter.UserID)
		}
		if filter.CategoryID != nil {
			builder = builder.Where("t.category_id = ?", *filter.CategoryID)
		}
		if filter.Type != nil {
			builder = builder.Where("t.type = ?", string(*filter.Type))
		}
		if filter.CurrencyCode != "" {
			builder = builder.Where("t.currency_code = ?", filter.CurrencyCode)
		}
		if filter.Source != nil {
			builder = builder.Where("t.source = ?", string(*filter.Source))
		}
		if filter.CreatedFrom != nil {
			builder = builder.Where(squirrel.GtOrEq{"t.created_at": *filter.CreatedFrom})
		}
		if filter.CreatedTo != nil {
			builder = builder.Where(squirrel.LtOrEq{"t.created_at": *filter.CreatedTo})
		}
	}

	// The default timeline is newest-first. ID makes rows with an identical
	// timestamp deterministic (and matches insertion order for BIGSERIAL ids).
	sql, args, err := builder.OrderBy("t.created_at DESC", "t.id DESC").ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(ctx.Context(), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		tr, err := scanTransactionRow(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, *tr)
	}
	return transactions, rows.Err()
}

func (r *PostgresTransactionRepository) Update(ctx domain.AppContext, tr domain.Transaction, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	amountNumeric, err := moneyToNumeric(tr.Amount)
	if err != nil {
		return err
	}
	sql, args, err := psql.Update("transactions").
		Set("type", string(tr.Type)).
		Set("amount", amountNumeric).
		Set("currency_code", tr.Currency.Code).
		Set("category_id", tr.Category.ID).
		Set("user_id", tr.User.ID).
		Set("source", string(tr.Source)).
		Set("name", tr.Name).
		Set("description", tr.Description).
		Where("id = ?", tr.ID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}

func (r *PostgresTransactionRepository) Delete(ctx domain.AppContext, tr domain.Transaction, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Delete("transactions").Where("id = ?", tr.ID).ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}
