package transaction_attachment

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/object_storage"
	"jww/internal/service/unit_of_work"
)

type OrphanReporter func(ctx domain.AppContext, attachmentID uuid.UUID, cleanupErr error)

type Service struct {
	repository   Repository
	storage      object_storage.Store
	reportOrphan OrphanReporter
}

func NewService(repository Repository, storage object_storage.Store, reportOrphan OrphanReporter) *Service {
	if reportOrphan == nil {
		reportOrphan = func(domain.AppContext, uuid.UUID, error) {}
	}
	return &Service{repository: repository, storage: storage, reportOrphan: reportOrphan}
}

func objectKey(id uuid.UUID) string { return id.String() }

func (s *Service) UploadDraft(ctx domain.AppContext, filename, mediaType string, body io.Reader) (*domain.TransactionAttachment, error) {
	return s.upload(ctx, nil, filename, mediaType, body)
}

func (s *Service) UploadForTransaction(ctx domain.AppContext, transactionID int64, filename, mediaType string, body io.Reader) (*domain.TransactionAttachment, error) {
	return s.upload(ctx, &transactionID, filename, mediaType, body)
}

func (s *Service) upload(ctx domain.AppContext, transactionID *int64, filename, mediaType string, body io.Reader) (*domain.TransactionAttachment, error) {
	id := uuid.New()
	attachment := domain.TransactionAttachment{
		ID:               id,
		TransactionID:    transactionID,
		OriginalFilename: filename,
		MediaType:        mediaType,
		StorageKey:       objectKey(id),
		CreatedAt:        time.Now().UTC(),
	}
	if err := attachment.Validate(); err != nil {
		return nil, err
	}

	counted := &countingReader{reader: body}
	stored, err := s.storage.Put(ctx, object_storage.Ref{Key: attachment.StorageKey}, counted, mediaType)
	if err != nil {
		return nil, domain.NewObjectStorageUnavailableError(err)
	}
	attachment.ByteSize = counted.n
	attachment.StorageVersion = stored.Version
	attachment.ETag = stored.ETag

	if err := s.repository.Create(ctx, attachment); err != nil {
		s.compensate(ctx, attachment)
		return nil, err
	}
	return &attachment, nil
}

func (s *Service) GetDraft(ctx domain.AppContext, attachmentID uuid.UUID) (*domain.TransactionAttachment, error) {
	attachment, err := s.repository.Find(ctx, Filter{ID: attachmentID, DraftOnly: true})
	if err != nil {
		return nil, err
	}
	if attachment == nil {
		return nil, domain.NewTransactionAttachmentNotFoundError(attachmentID)
	}
	return attachment, nil
}

func (s *Service) OpenDraft(ctx domain.AppContext, attachmentID uuid.UUID) (*domain.TransactionAttachment, object_storage.Stream, error) {
	attachment, err := s.GetDraft(ctx, attachmentID)
	if err != nil {
		return nil, object_storage.Stream{}, err
	}
	return s.open(ctx, *attachment)
}

func (s *Service) DeleteDraft(ctx domain.AppContext, attachmentID uuid.UUID) error {
	attachment, err := s.GetDraft(ctx, attachmentID)
	if err != nil {
		return err
	}
	return s.delete(ctx, *attachment)
}

func (s *Service) ListForTransaction(ctx domain.AppContext, transactionID int64, uow ...unit_of_work.UnitOfWork) ([]domain.TransactionAttachment, error) {
	return s.repository.ListByTransactionIDs(ctx, []int64{transactionID}, uow...)
}

func (s *Service) GetForTransaction(ctx domain.AppContext, transactionID int64, attachmentID uuid.UUID) (*domain.TransactionAttachment, error) {
	attachment, err := s.repository.Find(ctx, Filter{ID: attachmentID, TransactionID: &transactionID})
	if err != nil {
		return nil, err
	}
	if attachment == nil {
		return nil, domain.NewTransactionAttachmentNotFoundError(attachmentID)
	}
	return attachment, nil
}

func (s *Service) OpenForTransaction(ctx domain.AppContext, transactionID int64, attachmentID uuid.UUID) (*domain.TransactionAttachment, object_storage.Stream, error) {
	attachment, err := s.GetForTransaction(ctx, transactionID, attachmentID)
	if err != nil {
		return nil, object_storage.Stream{}, err
	}
	return s.open(ctx, *attachment)
}

func (s *Service) DeleteForTransaction(ctx domain.AppContext, transactionID int64, attachmentID uuid.UUID) error {
	attachment, err := s.GetForTransaction(ctx, transactionID, attachmentID)
	if err != nil {
		return err
	}
	return s.delete(ctx, *attachment)
}

func (s *Service) LockByIDs(ctx domain.AppContext, attachmentIDs []uuid.UUID, uow unit_of_work.UnitOfWork) ([]domain.TransactionAttachment, error) {
	return s.repository.LockByIDs(ctx, attachmentIDs, uow)
}

func (s *Service) Link(ctx domain.AppContext, transactionID int64, attachmentIDs []uuid.UUID, uow unit_of_work.UnitOfWork) error {
	return s.repository.LinkToTransaction(ctx, transactionID, attachmentIDs, uow)
}

func (s *Service) ListByTransactionIDs(ctx domain.AppContext, transactionIDs []int64, uow ...unit_of_work.UnitOfWork) ([]domain.TransactionAttachment, error) {
	if len(transactionIDs) == 0 {
		return nil, nil
	}
	return s.repository.ListByTransactionIDs(ctx, transactionIDs, uow...)
}

// DeleteAllForTransaction deletes every attachment as one logical resource:
// its object bytes first and its metadata row second. Callers must not split
// those responsibilities or rely on deletion of the parent transaction to
// clean up attachment metadata.
func (s *Service) DeleteAllForTransaction(ctx domain.AppContext, transactionID int64, uow ...unit_of_work.UnitOfWork) error {
	attachments, err := s.ListForTransaction(ctx, transactionID, uow...)
	if err != nil {
		return err
	}
	for _, attachment := range attachments {
		if err := s.delete(ctx, attachment, uow...); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) open(ctx domain.AppContext, attachment domain.TransactionAttachment) (*domain.TransactionAttachment, object_storage.Stream, error) {
	stream, err := s.storage.Open(ctx, object_storage.Ref{Key: attachment.StorageKey, Version: attachment.StorageVersion})
	if err != nil {
		if errors.Is(err, object_storage.ErrNotFound) {
			return nil, object_storage.Stream{}, domain.NewAttachmentIntegrityError(attachment.ID, err)
		}
		return nil, object_storage.Stream{}, domain.NewObjectStorageUnavailableError(err)
	}
	return &attachment, stream, nil
}

func (s *Service) delete(ctx domain.AppContext, attachment domain.TransactionAttachment, uow ...unit_of_work.UnitOfWork) error {
	if err := s.storage.Delete(ctx, object_storage.Ref{Key: attachment.StorageKey, Version: attachment.StorageVersion}); err != nil {
		return domain.NewObjectStorageUnavailableError(err)
	}
	return s.repository.Delete(ctx, attachment, uow...)
}

func (s *Service) compensate(ctx domain.AppContext, attachment domain.TransactionAttachment) {
	if err := s.storage.Delete(ctx, object_storage.Ref{Key: attachment.StorageKey, Version: attachment.StorageVersion}); err != nil {
		s.reportOrphan(ctx, attachment.ID, fmt.Errorf("deleting object after failed metadata write: %w", err))
	}
}

type countingReader struct {
	reader io.Reader
	n      int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.reader.Read(p)
	c.n += int64(n)
	return n, err
}
