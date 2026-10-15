package gateway

import (
	"errors"

	"github.com/jackc/pgx/v5"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

// PostgresCurrencyRepository implements service/currency.Repository.
type PostgresCurrencyRepository struct {
	db DBTX
}

func NewPostgresCurrencyRepository(db DBTX) *PostgresCurrencyRepository {
	return &PostgresCurrencyRepository{db: db}
}

func scanCurrencyRow(row pgx.Row) (*domain.Currency, error) {
	var c domain.Currency
	err := row.Scan(&c.Code, &c.Name, &c.IsBase)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *PostgresCurrencyRepository) Create(ctx domain.AppContext, currency domain.Currency, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Insert("currencies").
		Columns("code", "name", "is_base").
		Values(currency.Code, currency.Name, currency.IsBase).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}

func (r *PostgresCurrencyRepository) GetByCode(ctx domain.AppContext, code string, uow ...unit_of_work.UnitOfWork) (*domain.Currency, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select("code", "name", "is_base").
		From("currencies").
		Where("code = ?", code).
		ToSql()
	if err != nil {
		return nil, err
	}
	return scanCurrencyRow(db.QueryRow(ctx.Context(), sql, args...))
}

func (r *PostgresCurrencyRepository) List(ctx domain.AppContext, uow ...unit_of_work.UnitOfWork) ([]domain.Currency, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select("code", "name", "is_base").
		From("currencies").
		OrderBy("code").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(ctx.Context(), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currencies []domain.Currency
	for rows.Next() {
		c, err := scanCurrencyRow(rows)
		if err != nil {
			return nil, err
		}
		currencies = append(currencies, *c)
	}
	return currencies, rows.Err()
}

func (r *PostgresCurrencyRepository) Update(ctx domain.AppContext, currency domain.Currency, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Update("currencies").
		Set("name", currency.Name).
		Set("is_base", currency.IsBase).
		Where("code = ?", currency.Code).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}

func (r *PostgresCurrencyRepository) Delete(ctx domain.AppContext, currency domain.Currency, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Delete("currencies").Where("code = ?", currency.Code).ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}
