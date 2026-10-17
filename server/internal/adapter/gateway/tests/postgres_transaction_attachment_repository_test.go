//go:build integration

package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/gateway"
	"jww/internal/domain"
	"jww/internal/service/transaction_attachment"
	"jww/internal/service/unit_of_work"
	"jww/internal/tests/pgtest"
)

func newTestAttachment() domain.TransactionAttachment {
	id := uuid.New()
	return domain.TransactionAttachment{
		ID:               id,
		OriginalFilename: "receipt.pdf",
		MediaType:        "application/pdf",
		ByteSize:         42,
		StorageKey:       "transaction-attachments/" + id.String(),
		StorageVersion:   "v1",
		ETag:             "etag-v1",
		CreatedAt:        time.Now().UTC().Truncate(time.Microsecond),
	}
}

// createTransactionForAttachments inserts a transaction so attachment rows have
// a valid FK target, returning its id.
func createTransactionForAttachments(t *testing.T, appCtx domain.AppContext, uow unit_of_work.UnitOfWork, currencyCode string) int64 {
	t.Helper()
	fx := seedTransactionFixtures(t, appCtx, uow, currencyCode, domain.TransactionCategoryTypeExpense)
	txRepo := gateway.NewPostgresTransactionRepository(pool)
	tx, err := txRepo.Create(appCtx, domain.TransactionTypeExpense, domain.Money(100), fx.currency, fx.category, fx.user,
		domain.TransactionSourceManual, time.Now(), "attachment host", "", uow)
	require.NoError(t, err)
	return tx.ID
}

func TestPostgresTransactionAttachmentRepository_DraftRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionAttachmentRepository(pool)

	attachment := newTestAttachment()
	require.NoError(t, repo.Create(appCtx, attachment, uow))

	got, err := repo.GetByID(appCtx, attachment.ID, uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.TransactionID)
	assert.Equal(t, attachment.OriginalFilename, got.OriginalFilename)
	assert.Equal(t, attachment.MediaType, got.MediaType)
	assert.Equal(t, attachment.ByteSize, got.ByteSize)
	assert.Equal(t, attachment.StorageKey, got.StorageKey)
	assert.Equal(t, attachment.StorageVersion, got.StorageVersion)
	assert.True(t, attachment.CreatedAt.Equal(got.CreatedAt))
}

func TestPostgresTransactionAttachmentRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionAttachmentRepository(pool)

	got, err := repo.GetByID(appCtx, uuid.New(), uow)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresTransactionAttachmentRepository_LinkIsOneWay(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionAttachmentRepository(pool)

	txID := createTransactionForAttachments(t, appCtx, uow, "ZGA")
	attachment := newTestAttachment()
	require.NoError(t, repo.Create(appCtx, attachment, uow))

	require.NoError(t, repo.LinkToTransaction(appCtx, txID, []uuid.UUID{attachment.ID}, uow))

	// The identity-only lookup still finds a linked attachment.
	byID, err := repo.GetByID(appCtx, attachment.ID, uow)
	require.NoError(t, err)
	require.NotNil(t, byID)
	require.NotNil(t, byID.TransactionID)
	assert.Equal(t, txID, *byID.TransactionID)

	// Explicit filters enforce draft-only and parent-scoped access in SQL.
	draft, err := repo.Find(appCtx, transaction_attachment.Filter{ID: attachment.ID, DraftOnly: true}, uow)
	require.NoError(t, err)
	assert.Nil(t, draft)

	scoped, err := repo.Find(appCtx, transaction_attachment.Filter{ID: attachment.ID, TransactionID: &txID}, uow)
	require.NoError(t, err)
	require.NotNil(t, scoped)
	require.NotNil(t, scoped.TransactionID)
	assert.Equal(t, txID, *scoped.TransactionID)

	otherTxID := txID + 1
	wrongParent, err := repo.Find(appCtx, transaction_attachment.Filter{ID: attachment.ID, TransactionID: &otherTxID}, uow)
	require.NoError(t, err)
	assert.Nil(t, wrongParent)

	// A second link is rejected: the conditional update matches no row.
	assert.Error(t, repo.LinkToTransaction(appCtx, txID+1, []uuid.UUID{attachment.ID}, uow))
}

func TestPostgresTransactionAttachmentRepository_LockByIDs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionAttachmentRepository(pool)

	a1 := newTestAttachment()
	a2 := newTestAttachment()
	require.NoError(t, repo.Create(appCtx, a1, uow))
	require.NoError(t, repo.Create(appCtx, a2, uow))

	locked, err := repo.LockByIDs(appCtx, []uuid.UUID{a1.ID, a2.ID, uuid.New()}, uow)
	require.NoError(t, err)
	assert.Len(t, locked, 2, "only existing rows are returned")
}

func TestPostgresTransactionAttachmentRepository_ListOrdersDeterministically(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionAttachmentRepository(pool)

	txID := createTransactionForAttachments(t, appCtx, uow, "ZGB")

	earlier := newTestAttachment()
	earlier.CreatedAt = time.Now().UTC().Add(-time.Hour)
	earlier.TransactionID = &txID
	later := newTestAttachment()
	later.TransactionID = &txID
	require.NoError(t, repo.Create(appCtx, earlier, uow))
	require.NoError(t, repo.Create(appCtx, later, uow))

	list, err := repo.ListByTransactionIDs(appCtx, []int64{txID}, uow)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, earlier.ID, list[0].ID)
	assert.Equal(t, later.ID, list[1].ID)
}

func TestPostgresTransactionAttachmentRepository_StorageKeyUnique(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionAttachmentRepository(pool)

	first := newTestAttachment()
	require.NoError(t, repo.Create(appCtx, first, uow))

	duplicate := newTestAttachment()
	duplicate.StorageKey = first.StorageKey
	assert.Error(t, repo.Create(appCtx, duplicate, uow), "storage_key must be unique")
}

func TestPostgresTransactionAttachmentRepository_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionAttachmentRepository(pool)

	attachment := newTestAttachment()
	require.NoError(t, repo.Create(appCtx, attachment, uow))
	require.NoError(t, repo.Delete(appCtx, attachment, uow))

	got, err := repo.GetByID(appCtx, attachment.ID, uow)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresTransactionAttachmentRepository_TransactionDeleteIsRestricted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionAttachmentRepository(pool)
	txRepo := gateway.NewPostgresTransactionRepository(pool)

	fx := seedTransactionFixtures(t, appCtx, uow, "ZGC", domain.TransactionCategoryTypeExpense)
	tx, err := txRepo.Create(appCtx, domain.TransactionTypeExpense, domain.Money(100), fx.currency, fx.category, fx.user,
		domain.TransactionSourceManual, time.Now(), "attachment host", "", uow)
	require.NoError(t, err)

	attachment := newTestAttachment()
	attachment.TransactionID = &tx.ID
	require.NoError(t, repo.Create(appCtx, attachment, uow))

	require.Error(t, txRepo.Delete(appCtx, *tx, uow), "the attachment service must delete the attachment first")
}
