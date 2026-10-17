package gateway

import (
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"jww/internal/domain"
	"jww/internal/service/transaction_attachment"
	"jww/internal/service/unit_of_work"
)

// PostgresTransactionAttachmentRepository implements
// transaction_attachment.Repository.
type PostgresTransactionAttachmentRepository struct {
	db DBTX
}

func NewPostgresTransactionAttachmentRepository(db DBTX) *PostgresTransactionAttachmentRepository {
	return &PostgresTransactionAttachmentRepository{db: db}
}

const transactionAttachmentSelectColumns = `
	a.id, a.transaction_id, a.original_filename, a.media_type, a.byte_size,
	a.storage_key, a.storage_version, a.etag, a.created_at`

func scanTransactionAttachmentRow(row pgx.Row) (*domain.TransactionAttachment, error) {
	var a domain.TransactionAttachment
	var transactionID pgtype.Int8

	err := row.Scan(
		&a.ID, &transactionID, &a.OriginalFilename, &a.MediaType, &a.ByteSize,
		&a.StorageKey, &a.StorageVersion, &a.ETag, &a.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if transactionID.Valid {
		id := transactionID.Int64
		a.TransactionID = &id
	}
	return &a, nil
}

// nullableInt64 renders a nil *int64 as SQL NULL rather than a typed nil
// pointer pgx would reject.
func nullableInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func (r *PostgresTransactionAttachmentRepository) Create(ctx domain.AppContext, attachment domain.TransactionAttachment, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Insert("transaction_attachments").
		Columns("id", "transaction_id", "original_filename", "media_type", "byte_size", "storage_key", "storage_version", "etag", "created_at").
		Values(
			attachment.ID, nullableInt64(attachment.TransactionID), attachment.OriginalFilename,
			attachment.MediaType, attachment.ByteSize, attachment.StorageKey,
			attachment.StorageVersion, attachment.ETag, attachment.CreatedAt,
		).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}

func (r *PostgresTransactionAttachmentRepository) GetByID(ctx domain.AppContext, attachmentID uuid.UUID, uow ...unit_of_work.UnitOfWork) (*domain.TransactionAttachment, error) {
	return r.Find(ctx, transaction_attachment.Filter{ID: attachmentID}, uow...)
}

// Find applies explicit attachment access constraints in SQL.
func (r *PostgresTransactionAttachmentRepository) Find(ctx domain.AppContext, filter transaction_attachment.Filter, uow ...unit_of_work.UnitOfWork) (*domain.TransactionAttachment, error) {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	builder := psql.Select(transactionAttachmentSelectColumns[1:]).
		From("transaction_attachments a").
		Where("a.id = ?", filter.ID)
	if filter.DraftOnly {
		builder = builder.Where("a.transaction_id IS NULL")
	}
	if filter.TransactionID != nil {
		builder = builder.Where("a.transaction_id = ?", *filter.TransactionID)
	}
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	return scanTransactionAttachmentRow(db.QueryRow(ctx.Context(), sql, args...))
}

// LockByIDs selects and locks the requested rows in a stable id order, so two
// concurrent claims take locks in the same sequence and cannot deadlock.
func (r *PostgresTransactionAttachmentRepository) LockByIDs(ctx domain.AppContext, attachmentIDs []uuid.UUID, uow ...unit_of_work.UnitOfWork) ([]domain.TransactionAttachment, error) {
	if len(attachmentIDs) == 0 {
		return nil, nil
	}
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select(transactionAttachmentSelectColumns[1:]).
		From("transaction_attachments a").
		Where(squirrel.Eq{"a.id": attachmentIDs}).
		OrderBy("a.id").
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(ctx.Context(), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []domain.TransactionAttachment
	for rows.Next() {
		a, err := scanTransactionAttachmentRow(rows)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, *a)
	}
	return attachments, rows.Err()
}

// LinkToTransaction performs the one-way NULL -> transactionID transition. The
// conditional WHERE means an already-linked row is never reassigned; a mismatch
// between rows affected and rows requested is surfaced rather than silently
// ignored.
func (r *PostgresTransactionAttachmentRepository) LinkToTransaction(ctx domain.AppContext, transactionID int64, attachmentIDs []uuid.UUID, uow ...unit_of_work.UnitOfWork) error {
	if len(attachmentIDs) == 0 {
		return nil
	}
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Update("transaction_attachments").
		Set("transaction_id", transactionID).
		Where(squirrel.Eq{"id": attachmentIDs}).
		Where("transaction_id IS NULL").
		ToSql()
	if err != nil {
		return err
	}
	tag, err := db.Exec(ctx.Context(), sql, args...)
	if err != nil {
		return err
	}
	if n := int(tag.RowsAffected()); n != len(attachmentIDs) {
		return fmt.Errorf("gateway: linked %d of %d requested attachments", n, len(attachmentIDs))
	}
	return nil
}

func (r *PostgresTransactionAttachmentRepository) ListByTransactionIDs(ctx domain.AppContext, transactionIDs []int64, uow ...unit_of_work.UnitOfWork) ([]domain.TransactionAttachment, error) {
	if len(transactionIDs) == 0 {
		return nil, nil
	}
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return nil, err
	}
	sql, args, err := psql.Select(transactionAttachmentSelectColumns[1:]).
		From("transaction_attachments a").
		Where(squirrel.Eq{"a.transaction_id": transactionIDs}).
		OrderBy("a.created_at", "a.id").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(ctx.Context(), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []domain.TransactionAttachment
	for rows.Next() {
		a, err := scanTransactionAttachmentRow(rows)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, *a)
	}
	return attachments, rows.Err()
}

func (r *PostgresTransactionAttachmentRepository) Delete(ctx domain.AppContext, attachment domain.TransactionAttachment, uow ...unit_of_work.UnitOfWork) error {
	db, err := dbFromUnitOfWork(r.db, uow...)
	if err != nil {
		return err
	}
	sql, args, err := psql.Delete("transaction_attachments").Where("id = ?", attachment.ID).ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx.Context(), sql, args...)
	return err
}
