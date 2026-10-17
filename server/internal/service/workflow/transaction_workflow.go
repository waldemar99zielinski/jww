package workflow

import (
	"fmt"
	"io"
	"sort"

	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/currency"
	"jww/internal/service/transaction"
	"jww/internal/service/transaction_attachment"
	"jww/internal/service/transaction_category"
	"jww/internal/service/unit_of_work"
	"jww/internal/service/user"
)

// TransactionWorkflow resolves the currency/category/user a transaction
// references before handing off to transaction.Service — the orchestration
// that create_transaction_handler.go and list_transactions_handler.go used
// to do inline, composing four services directly in the HTTP layer. It also
// owns the attachment side of a transaction's lifecycle: create-and-link,
// attach-later, hydration of reads, and object cleanup on delete.
type TransactionWorkflow struct {
	currencyService    *currency.Service
	categoryService    *transaction_category.Service
	userService        *user.Service
	transactionService *transaction.Service
	attachmentService  *transaction_attachment.Service
	uowManager         unit_of_work.Manager
}

func NewTransactionWorkflow(
	currencyService *currency.Service,
	categoryService *transaction_category.Service,
	userService *user.Service,
	transactionService *transaction.Service,
	attachmentService *transaction_attachment.Service,
	uowManager unit_of_work.Manager,
) *TransactionWorkflow {
	return &TransactionWorkflow{
		currencyService:    currencyService,
		categoryService:    categoryService,
		userService:        userService,
		transactionService: transactionService,
		attachmentService:  attachmentService,
		uowManager:         uowManager,
	}
}

// Create resolves references and, when attachmentIDs is non-empty, locks and
// links those drafts in the same unit of work as the insert — so the
// transaction and every link commit or roll back together, with no S3 work.
func (w *TransactionWorkflow) Create(
	ctx domain.AppContext,
	transactionType domain.TransactionType,
	amount domain.Money,
	currencyCode string,
	categoryID uuid.UUID,
	userID uuid.UUID,
	source domain.TransactionSource,
	name string,
	description string,
	attachmentIDs []uuid.UUID,
) (*domain.Transaction, error) {
	uow, err := w.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	curr, err := w.currencyService.GetByCode(ctx, currencyCode, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	category, err := w.categoryService.GetByID(ctx, categoryID, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	u, err := w.userService.GetByID(ctx, userID, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}

	var attachments []domain.TransactionAttachment
	if len(attachmentIDs) > 0 {
		locked, err := w.attachmentService.LockByIDs(ctx, attachmentIDs, uow)
		if err != nil {
			_ = uow.Restore(ctx)
			return nil, err
		}
		if err := claimable(attachmentIDs, locked); err != nil {
			_ = uow.Restore(ctx)
			return nil, err
		}
		attachments = locked
	}

	created, err := w.transactionService.Create(ctx, transactionType, amount, *curr, *category, *u, source, name, description, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}

	if len(attachmentIDs) > 0 {
		if err := w.attachmentService.Link(ctx, created.ID, attachmentIDs, uow); err != nil {
			_ = uow.Restore(ctx)
			return nil, err
		}
		created.Attachments = linkedAs(attachments, created.ID)
	} else {
		created.Attachments = []domain.TransactionAttachment{}
	}

	if err := uow.Finish(ctx); err != nil {
		return nil, err
	}
	return created, nil
}

// Attach links an existing unattached draft to an existing transaction. It is
// idempotent for the same parent and a conflict for a different one.
func (w *TransactionWorkflow) Attach(ctx domain.AppContext, transactionID int64, attachmentID uuid.UUID) error {
	uow, err := w.uowManager.Create(ctx)
	if err != nil {
		return err
	}
	if _, err := w.transactionService.GetByID(ctx, transactionID, uow); err != nil {
		_ = uow.Restore(ctx)
		return err
	}
	locked, err := w.attachmentService.LockByIDs(ctx, []uuid.UUID{attachmentID}, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return err
	}
	if len(locked) == 0 {
		_ = uow.Restore(ctx)
		return domain.NewTransactionAttachmentNotFoundError(attachmentID)
	}
	if existing := locked[0].TransactionID; existing != nil {
		if *existing == transactionID {
			return uow.Finish(ctx) // already linked here: no-op
		}
		_ = uow.Restore(ctx)
		return domain.NewAttachmentAlreadyLinkedError(attachmentID)
	}
	if err := w.attachmentService.Link(ctx, transactionID, []uuid.UUID{attachmentID}, uow); err != nil {
		_ = uow.Restore(ctx)
		return err
	}
	return uow.Finish(ctx)
}

// Get loads one transaction together with its attachment summaries.
func (w *TransactionWorkflow) Get(ctx domain.AppContext, transactionID int64) (*domain.Transaction, error) {
	uow, err := w.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	t, err := w.transactionService.GetByID(ctx, transactionID, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	attachments, err := w.attachmentService.ListForTransaction(ctx, transactionID, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	t.Attachments = attachments
	if t.Attachments == nil {
		t.Attachments = []domain.TransactionAttachment{}
	}
	if err := uow.Finish(ctx); err != nil {
		return nil, err
	}
	return t, nil
}

// Delete asks the attachment service to delete every attachment in full before
// deleting the transaction. On the first attachment failure it stops and
// preserves the transaction for a retry.
func (w *TransactionWorkflow) Delete(ctx domain.AppContext, transactionID int64) error {
	uow, err := w.uowManager.Create(ctx)
	if err != nil {
		return err
	}
	t, err := w.transactionService.GetByID(ctx, transactionID, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return err
	}
	if err := w.attachmentService.DeleteAllForTransaction(ctx, transactionID, uow); err != nil {
		_ = uow.Restore(ctx)
		return err
	}
	if err := w.transactionService.Delete(ctx, *t, uow); err != nil {
		_ = uow.Restore(ctx)
		return err
	}
	return uow.Finish(ctx)
}

// UploadAttachment is the transaction-first upload path: the metadata row is
// created already linked to transactionID. The parent must exist first.
func (w *TransactionWorkflow) UploadAttachment(ctx domain.AppContext, transactionID int64, filename, mediaType string, body io.Reader) (*domain.TransactionAttachment, error) {
	if _, err := w.transactionService.GetByID(ctx, transactionID); err != nil {
		return nil, err
	}
	return w.attachmentService.UploadForTransaction(ctx, transactionID, filename, mediaType, body)
}

func (w *TransactionWorkflow) List(ctx domain.AppContext, filter transaction.Filter) ([]domain.Transaction, error) {
	if filter.UserID != nil {
		if _, err := w.userService.GetByID(ctx, *filter.UserID); err != nil {
			return nil, err
		}
	}
	if filter.CategoryID != nil {
		category, err := w.categoryService.GetByID(ctx, *filter.CategoryID)
		if err != nil {
			return nil, err
		}
		if filter.Type != nil && category.Type != domain.TransactionCategoryType(*filter.Type) {
			return nil, fmt.Errorf("category type must match the filter's transaction type")
		}
	}
	if filter.CurrencyCode != "" {
		if _, err := w.currencyService.GetByCode(ctx, filter.CurrencyCode); err != nil {
			return nil, err
		}
	}
	transactions, err := w.transactionService.List(ctx, &filter)
	if err != nil {
		return nil, err
	}
	if err := w.hydrate(ctx, transactions); err != nil {
		return nil, err
	}
	return transactions, nil
}

// hydrate batch-loads attachments for every returned transaction in one query
// and groups them in memory
func (w *TransactionWorkflow) hydrate(ctx domain.AppContext, transactions []domain.Transaction) error {
	ids := make([]int64, len(transactions))
	for i := range transactions {
		ids[i] = transactions[i].ID
	}
	attachments, err := w.attachmentService.ListByTransactionIDs(ctx, ids)
	if err != nil {
		return err
	}
	byTransaction := make(map[int64][]domain.TransactionAttachment, len(transactions))
	for _, attachment := range attachments {
		if attachment.TransactionID == nil {
			continue
		}
		byTransaction[*attachment.TransactionID] = append(byTransaction[*attachment.TransactionID], attachment)
	}
	for i := range transactions {
		group := byTransaction[transactions[i].ID]
		if group == nil {
			group = []domain.TransactionAttachment{}
		}
		transactions[i].Attachments = group
	}
	return nil
}

// claimable verifies every requested id exists (else not found) and is still an
// unlinked draft (else conflict). Ids are de-duplicated by the request schema;
// duplicate detection here is defensive.
func claimable(requested []uuid.UUID, found []domain.TransactionAttachment) error {
	byID := make(map[uuid.UUID]domain.TransactionAttachment, len(found))
	for _, attachment := range found {
		byID[attachment.ID] = attachment
	}
	seen := make(map[uuid.UUID]struct{}, len(requested))
	for _, id := range requested {
		if _, dup := seen[id]; dup {
			return domain.NewValidationError("attachment_ids must not contain duplicate ids")
		}
		seen[id] = struct{}{}
		attachment, ok := byID[id]
		if !ok {
			return domain.NewTransactionAttachmentNotFoundError(id)
		}
		if attachment.TransactionID != nil {
			return domain.NewAttachmentAlreadyLinkedError(id)
		}
	}
	return nil
}

// linkedAs sorts the locked attachments into the same deterministic order the
// read paths use (created_at, then id) and stamps them with the transaction id
// they were just linked to, so the create response reflects the new state.
func linkedAs(attachments []domain.TransactionAttachment, transactionID int64) []domain.TransactionAttachment {
	out := make([]domain.TransactionAttachment, len(attachments))
	copy(out, attachments)
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].CreatedAt.Before(out[j].CreatedAt)
		}
		return out[i].ID.String() < out[j].ID.String()
	})
	for i := range out {
		id := transactionID
		out[i].TransactionID = &id
	}
	return out
}
