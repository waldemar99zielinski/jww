package gateway

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

// PostgresWebAuthnCredentialRepository implements
// service/webauthn_credential.Repository.
type PostgresWebAuthnCredentialRepository struct {
	db DBTX
}

func NewPostgresWebAuthnCredentialRepository(db DBTX) *PostgresWebAuthnCredentialRepository {
	return &PostgresWebAuthnCredentialRepository{db: db}
}

const webauthnCredentialSelectColumns = `
	id, user_id, name, credential_id, public_key, attestation_type, aaguid,
	sign_count, transports, backup_eligible, created_at, last_used_at`

func scanWebAuthnCredentialRow(row pgx.Row) (*domain.WebAuthnCredential, error) {
	var c domain.WebAuthnCredential
	var signCount int64
	err := row.Scan(
		&c.ID, &c.UserID, &c.Name, &c.CredentialID, &c.PublicKey, &c.AttestationType, &c.AAGUID,
		&signCount, &c.Transports, &c.BackupEligible, &c.CreatedAt, &c.LastUsedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.SignCount = uint32(signCount)
	return &c, nil
}

func (r *PostgresWebAuthnCredentialRepository) Create(ctx domain.AppContext, credential domain.WebAuthnCredential, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Insert("webauthn_credentials").
		Columns("id", "user_id", "name", "credential_id", "public_key", "attestation_type", "aaguid", "sign_count", "transports", "backup_eligible", "created_at", "last_used_at").
		Values(
			credential.ID, credential.UserID, credential.Name, credential.CredentialID, credential.PublicKey,
			credential.AttestationType, credential.AAGUID, credential.SignCount, credential.Transports,
			credential.BackupEligible, credential.CreatedAt, credential.LastUsedAt,
		).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}

func (r *PostgresWebAuthnCredentialRepository) GetByID(ctx domain.AppContext, id uuid.UUID, uow ...unit_of_work.UnitOfWork) (*domain.WebAuthnCredential, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select(webauthnCredentialSelectColumns[1:]).
		From("webauthn_credentials").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return nil, err
	}
	return scanWebAuthnCredentialRow(db.QueryRow(ctx.Context(), sql, args...))
}

func (r *PostgresWebAuthnCredentialRepository) GetByCredentialID(ctx domain.AppContext, credentialID []byte, uow ...unit_of_work.UnitOfWork) (*domain.WebAuthnCredential, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select(webauthnCredentialSelectColumns[1:]).
		From("webauthn_credentials").
		Where("credential_id = ?", credentialID).
		ToSql()
	if err != nil {
		return nil, err
	}
	return scanWebAuthnCredentialRow(db.QueryRow(ctx.Context(), sql, args...))
}

func (r *PostgresWebAuthnCredentialRepository) GetByUserID(ctx domain.AppContext, userID uuid.UUID, uow ...unit_of_work.UnitOfWork) ([]domain.WebAuthnCredential, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select(webauthnCredentialSelectColumns[1:]).
		From("webauthn_credentials").
		Where("user_id = ?", userID).
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

	var credentials []domain.WebAuthnCredential
	for rows.Next() {
		c, err := scanWebAuthnCredentialRow(rows)
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, *c)
	}
	return credentials, rows.Err()
}

func (r *PostgresWebAuthnCredentialRepository) UpdateSignCount(ctx domain.AppContext, id uuid.UUID, signCount uint32, lastUsedAt time.Time, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Update("webauthn_credentials").
		Set("sign_count", signCount).
		Set("last_used_at", lastUsedAt).
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}

func (r *PostgresWebAuthnCredentialRepository) Delete(ctx domain.AppContext, id uuid.UUID, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Delete("webauthn_credentials").Where("id = ?", id).ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}
