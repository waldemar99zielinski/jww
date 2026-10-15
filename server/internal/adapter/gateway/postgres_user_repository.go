package gateway

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

// PostgresUserRepository implements both service/user.Repository and
// service/auth.UserRepository — two separate, independently-owned
// interfaces (the auth one is a strict subset) that happen to be
// satisfiable by the same concrete rows, satisfied implicitly here by
// Go's structural typing rather than declared explicitly.
type PostgresUserRepository struct {
	db DBTX
}

func NewPostgresUserRepository(db DBTX) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func scanUserRow(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) Create(ctx domain.AppContext, user domain.User, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Insert("users").
		Columns("id", "username", "password_hash", "display_name", "created_at").
		Values(user.ID, user.Username, user.PasswordHash, user.DisplayName, user.CreatedAt).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}

func (r *PostgresUserRepository) GetByID(ctx domain.AppContext, userID uuid.UUID, uow ...unit_of_work.UnitOfWork) (*domain.User, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select("id", "username", "password_hash", "display_name", "created_at").
		From("users").
		Where("id = ?", userID).
		ToSql()
	if err != nil {
		return nil, err
	}
	return scanUserRow(db.QueryRow(ctx.Context(), sql, args...))
}

func (r *PostgresUserRepository) GetByUsername(ctx domain.AppContext, username string, uow ...unit_of_work.UnitOfWork) (*domain.User, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select("id", "username", "password_hash", "display_name", "created_at").
		From("users").
		Where("username = ?", username).
		ToSql()
	if err != nil {
		return nil, err
	}
	return scanUserRow(db.QueryRow(ctx.Context(), sql, args...))
}

func (r *PostgresUserRepository) List(ctx domain.AppContext, uow ...unit_of_work.UnitOfWork) ([]domain.User, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select("id", "username", "password_hash", "display_name", "created_at").
		From("users").
		OrderBy("created_at").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(ctx.Context(), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		user, err := scanUserRow(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}
	return users, rows.Err()
}

func (r *PostgresUserRepository) Update(ctx domain.AppContext, user domain.User, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Update("users").
		Set("username", user.Username).
		Set("password_hash", user.PasswordHash).
		Set("display_name", user.DisplayName).
		Where("id = ?", user.ID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}

func (r *PostgresUserRepository) Delete(ctx domain.AppContext, user domain.User, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Delete("users").Where("id = ?", user.ID).ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}
