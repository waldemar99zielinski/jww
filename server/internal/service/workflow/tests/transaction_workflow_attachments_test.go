package workflow_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/transaction"
)

// draftAttachment builds a valid unattached attachment for the fakes.
func draftAttachment() domain.TransactionAttachment {
	id := uuid.New()
	return domain.TransactionAttachment{
		ID:               id,
		OriginalFilename: "receipt.pdf",
		MediaType:        "application/pdf",
		ByteSize:         11,
		StorageKey:       id.String(),
		StorageVersion:   "v1",
		CreatedAt:        time.Now().UTC(),
	}
}

func linkedAttachment(transactionID int64) domain.TransactionAttachment {
	a := draftAttachment()
	a.TransactionID = &transactionID
	return a
}

func expenseFixture() *transactionWorkflowFixture {
	category := groceries()
	user := julcia()
	return newTransactionWorkflowFixture([]domain.Currency{usd}, []domain.TransactionCategory{category}, []domain.User{user})
}

func TestTransactionWorkflow_Create_WithAttachments(t *testing.T) {
	t.Run("missing attachment rolls back before inserting the transaction", func(t *testing.T) {
		f := expenseFixture()
		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD",
			uuid.New(), uuid.New(), domain.TransactionSourceManual, "Shop", "", nil, []uuid.UUID{uuid.New()})

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.attachmentRepo.linkCalls)
		assert.Zero(t, f.transactionRepo.createCalled)
	})
}

// TestTransactionWorkflow_CreateAndLink uses the fixture's actual category and
// user so resolution succeeds and exercises the link path end to end.
func TestTransactionWorkflow_CreateAndLink(t *testing.T) {
	category := groceries()
	user := julcia()
	newFixture := func() *transactionWorkflowFixture {
		return newTransactionWorkflowFixture([]domain.Currency{usd}, []domain.TransactionCategory{category}, []domain.User{user})
	}

	t.Run("links an unattached draft and stamps the response", func(t *testing.T) {
		f := newFixture()
		attachment := draftAttachment()
		f.attachmentRepo.byID[attachment.ID] = attachment

		created, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD",
			category.ID, user.ID, domain.TransactionSourceManual, "Shop", "", nil, []uuid.UUID{attachment.ID})

		require.NoError(t, err)
		require.Len(t, created.Attachments, 1)
		require.NotNil(t, created.Attachments[0].TransactionID)
		assert.Equal(t, created.ID, *created.Attachments[0].TransactionID)
		assert.Equal(t, 1, f.attachmentRepo.linkCalls)
		assert.True(t, f.uowManager.last().finished)
	})

	t.Run("without attachment ids the response carries an empty, non-nil slice", func(t *testing.T) {
		f := newFixture()
		created, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD",
			category.ID, user.ID, domain.TransactionSourceManual, "Shop", "", nil, nil)

		require.NoError(t, err)
		assert.NotNil(t, created.Attachments)
		assert.Empty(t, created.Attachments)
	})

	t.Run("missing draft is not found and nothing is inserted", func(t *testing.T) {
		f := newFixture()
		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD",
			category.ID, user.ID, domain.TransactionSourceManual, "Shop", "", nil, []uuid.UUID{uuid.New()})

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.attachmentRepo.linkCalls)
		assert.Zero(t, f.transactionRepo.createCalled)
	})

	t.Run("already-linked draft is a conflict", func(t *testing.T) {
		f := newFixture()
		attachment := linkedAttachment(99)
		f.attachmentRepo.byID[attachment.ID] = attachment

		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD",
			category.ID, user.ID, domain.TransactionSourceManual, "Shop", "", nil, []uuid.UUID{attachment.ID})

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.createCalled)
	})

	t.Run("duplicate ids are a validation failure", func(t *testing.T) {
		f := newFixture()
		attachment := draftAttachment()
		f.attachmentRepo.byID[attachment.ID] = attachment

		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD",
			category.ID, user.ID, domain.TransactionSourceManual, "Shop", "", nil, []uuid.UUID{attachment.ID, attachment.ID})

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.createCalled)
	})
}

func TestTransactionWorkflow_Attach(t *testing.T) {
	txID := int64(7)
	newFixture := func() *transactionWorkflowFixture {
		f := expenseFixture()
		f.transactionRepo.getResult = &domain.Transaction{ID: txID}
		return f
	}

	t.Run("links an unattached draft", func(t *testing.T) {
		f := newFixture()
		attachment := draftAttachment()
		f.attachmentRepo.byID[attachment.ID] = attachment

		require.NoError(t, f.workflow.Attach(newAppContext(), txID, attachment.ID))
		assert.Equal(t, 1, f.attachmentRepo.linkCalls)
		assert.True(t, f.uowManager.last().finished)
	})

	t.Run("same parent is an idempotent no-op", func(t *testing.T) {
		f := newFixture()
		attachment := linkedAttachment(txID)
		f.attachmentRepo.byID[attachment.ID] = attachment

		require.NoError(t, f.workflow.Attach(newAppContext(), txID, attachment.ID))
		assert.Zero(t, f.attachmentRepo.linkCalls)
		assert.True(t, f.uowManager.last().finished)
	})

	t.Run("another parent is a conflict", func(t *testing.T) {
		f := newFixture()
		attachment := linkedAttachment(99)
		f.attachmentRepo.byID[attachment.ID] = attachment

		err := f.workflow.Attach(newAppContext(), txID, attachment.ID)
		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.attachmentRepo.linkCalls)
	})

	t.Run("missing attachment is not found", func(t *testing.T) {
		f := newFixture()
		err := f.workflow.Attach(newAppContext(), txID, uuid.New())
		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
	})

	t.Run("missing parent rolls back", func(t *testing.T) {
		f := expenseFixture() // getResult nil
		err := f.workflow.Attach(newAppContext(), txID, uuid.New())
		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
	})
}

func TestTransactionWorkflow_Get_HydratesAttachments(t *testing.T) {
	f := expenseFixture()
	f.transactionRepo.getResult = &domain.Transaction{ID: 3}
	f.attachmentRepo.byID[uuid.New()] = linkedAttachment(3)
	f.attachmentRepo.byID[uuid.New()] = linkedAttachment(3)

	got, err := f.workflow.Get(newAppContext(), 3)
	require.NoError(t, err)
	assert.Len(t, got.Attachments, 2)
	assert.True(t, f.uowManager.last().finished)
}

func TestTransactionWorkflow_List_HydratesInOneBatch(t *testing.T) {
	f := expenseFixture()
	f.transactionRepo.listResult = []domain.Transaction{{ID: 1}, {ID: 2}}
	f.attachmentRepo.byID[uuid.New()] = linkedAttachment(1)
	f.attachmentRepo.byID[uuid.New()] = linkedAttachment(2)

	got, err := f.workflow.List(newAppContext(), transaction.Filter{})
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Nil(t, f.uowManager.last())
	for _, t2 := range got {
		assert.Len(t, t2.Attachments, 1)
	}
}

func TestTransactionWorkflow_Delete_RemovesAttachmentsBeforeTransaction(t *testing.T) {
	t.Run("success explicitly deletes every attachment then the transaction", func(t *testing.T) {
		f := expenseFixture()
		f.transactionRepo.getResult = &domain.Transaction{ID: 1}
		a1 := linkedAttachment(1)
		a2 := linkedAttachment(1)
		f.attachmentRepo.byID[a1.ID] = a1
		f.attachmentRepo.byID[a2.ID] = a2
		f.attachmentStorage.objects[a1.StorageKey] = []byte("one")
		f.attachmentStorage.objects[a2.StorageKey] = []byte("two")

		require.NoError(t, f.workflow.Delete(newAppContext(), 1))
		assert.ElementsMatch(t, []string{a1.StorageKey, a2.StorageKey}, f.attachmentStorage.deleted)
		assert.NotContains(t, f.attachmentRepo.byID, a1.ID)
		assert.NotContains(t, f.attachmentRepo.byID, a2.ID)
		assert.Len(t, f.transactionRepo.deleted, 1)
		uow := f.uowManager.last()
		require.NotNil(t, uow)
		assert.True(t, uow.finished)
		assert.False(t, uow.restored)
		assert.Same(t, uow, f.transactionRepo.deleteUOW)
		require.Len(t, f.attachmentRepo.deleteUOWs, 2)
		for _, attachmentUOW := range f.attachmentRepo.deleteUOWs {
			assert.Same(t, uow, attachmentUOW)
		}
	})

	t.Run("a storage failure preserves the transaction and its rows", func(t *testing.T) {
		f := expenseFixture()
		f.transactionRepo.getResult = &domain.Transaction{ID: 1}
		a1 := linkedAttachment(1)
		f.attachmentRepo.byID[a1.ID] = a1
		f.attachmentStorage.deleteErr = assert.AnError

		err := f.workflow.Delete(newAppContext(), 1)
		require.Error(t, err)
		assert.Empty(t, f.transactionRepo.deleted)
		assert.True(t, f.uowManager.last().restored)
		assert.False(t, f.uowManager.last().finished)
		_, stillThere := f.attachmentRepo.byID[a1.ID]
		assert.True(t, stillThere)
	})
}

func TestTransactionWorkflow_UploadAttachment(t *testing.T) {
	t.Run("parent must exist", func(t *testing.T) {
		f := expenseFixture()
		_, err := f.workflow.UploadAttachment(newAppContext(), 5, "a.pdf", "application/pdf", strings.NewReader("data"))
		require.Error(t, err)
		assert.Zero(t, f.attachmentStorage.putCalls)
	})

	t.Run("creates metadata already linked", func(t *testing.T) {
		f := expenseFixture()
		f.transactionRepo.getResult = &domain.Transaction{ID: 5}

		attachment, err := f.workflow.UploadAttachment(newAppContext(), 5, "a.pdf", "application/pdf", strings.NewReader("data"))
		require.NoError(t, err)
		require.NotNil(t, attachment.TransactionID)
		assert.Equal(t, int64(5), *attachment.TransactionID)
		assert.Equal(t, int64(4), attachment.ByteSize)
		assert.Len(t, f.attachmentRepo.created, 1)
	})
}
