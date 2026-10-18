package transaction_attachment

import (
	"errors"
	"fmt"
	"mime"
	"strings"

	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/object_storage"
)

type GrantPurpose string

const (
	GrantPurposePreview  GrantPurpose = "preview"
	GrantPurposeDownload GrantPurpose = "download"
)

func ParseGrantPurpose(value string) (GrantPurpose, error) {
	purpose := GrantPurpose(value)
	switch purpose {
	case GrantPurposePreview, GrantPurposeDownload:
		return purpose, nil
	default:
		return "", fmt.Errorf("purpose must be preview or download")
	}
}

type GrantClaims struct {
	AttachmentID   uuid.UUID
	TransactionID  *int64
	StorageVersion string
	Purpose        GrantPurpose
}

type GrantTokenService interface {
	Issue(claims GrantClaims) (string, error)
	Verify(token string) (GrantClaims, error)
}

// GrantedAttachment is the verified, current attachment represented by a
// grant. Its fields stay private so callers cannot manufacture a grant.
type GrantedAttachment struct {
	attachment domain.TransactionAttachment
	purpose    GrantPurpose
}

func (g GrantedAttachment) Attachment() domain.TransactionAttachment { return g.attachment }
func (g GrantedAttachment) Purpose() GrantPurpose                    { return g.purpose }

func (s *Service) IssueDraftGrant(ctx domain.AppContext, attachmentID uuid.UUID, purpose GrantPurpose) (string, error) {
	attachment, err := s.GetDraft(ctx, attachmentID)
	if err != nil {
		return "", err
	}
	return s.issueGrant(*attachment, purpose)
}

func (s *Service) issueGrant(attachment domain.TransactionAttachment, purpose GrantPurpose) (string, error) {
	if s.grantTokens == nil {
		return "", fmt.Errorf("transaction_attachment: grant token service is not configured")
	}
	if _, err := ParseGrantPurpose(string(purpose)); err != nil {
		return "", domain.NewValidationError(err.Error())
	}
	if purpose == GrantPurposePreview && !previewableMediaType(attachment.MediaType) {
		return "", domain.NewValidationError("attachment media type cannot be previewed")
	}
	return s.grantTokens.Issue(GrantClaims{
		AttachmentID:   attachment.ID,
		TransactionID:  attachment.TransactionID,
		StorageVersion: attachment.StorageVersion,
		Purpose:        purpose,
	})
}

// IssuePreviewGrants issues grants for the previewable attachments in an
// already-authorized collection. Non-previewable attachments are omitted so
// transport schemas can represent them with a null preview_grant.
func (s *Service) IssuePreviewGrants(attachments []domain.TransactionAttachment) (map[uuid.UUID]string, error) {
	grants := make(map[uuid.UUID]string, len(attachments))
	for _, attachment := range attachments {
		if !previewableMediaType(attachment.MediaType) {
			continue
		}
		grant, err := s.issueGrant(attachment, GrantPurposePreview)
		if err != nil {
			return nil, err
		}
		grants[attachment.ID] = grant
	}
	return grants, nil
}

func (s *Service) ResolveGrant(ctx domain.AppContext, token string) (*GrantedAttachment, error) {
	if s.grantTokens == nil {
		return nil, fmt.Errorf("transaction_attachment: grant token service is not configured")
	}
	claims, err := s.grantTokens.Verify(token)
	if err != nil {
		return nil, domain.NewAuthenticationError()
	}

	var attachment *domain.TransactionAttachment
	if claims.TransactionID == nil {
		attachment, err = s.GetDraft(ctx, claims.AttachmentID)
	} else {
		attachment, err = s.GetForTransaction(ctx, *claims.TransactionID, claims.AttachmentID)
	}
	if err != nil {
		return nil, err
	}
	if attachment.StorageVersion != claims.StorageVersion {
		return nil, domain.NewAuthenticationError()
	}
	if claims.Purpose == GrantPurposePreview && !previewableMediaType(attachment.MediaType) {
		return nil, domain.NewAuthenticationError()
	}
	return &GrantedAttachment{attachment: *attachment, purpose: claims.Purpose}, nil
}

func (s *Service) OpenGranted(ctx domain.AppContext, granted GrantedAttachment, byteRange *object_storage.ByteRange) (object_storage.Stream, error) {
	attachment := granted.attachment
	ref := object_storage.Ref{Key: attachment.StorageKey, Version: attachment.StorageVersion}
	var (
		stream object_storage.Stream
		err    error
	)
	if byteRange == nil {
		stream, err = s.storage.Open(ctx, ref)
	} else {
		if byteRange.Start < 0 || byteRange.End < byteRange.Start || byteRange.End >= attachment.ByteSize {
			return object_storage.Stream{}, fmt.Errorf("transaction_attachment: invalid byte range")
		}
		stream, err = s.storage.OpenRange(ctx, ref, *byteRange)
	}
	if err != nil {
		if errors.Is(err, object_storage.ErrNotFound) {
			return object_storage.Stream{}, domain.NewAttachmentIntegrityError(attachment.ID, err)
		}
		return object_storage.Stream{}, domain.NewObjectStorageUnavailableError(err)
	}
	return stream, nil
}

func previewableMediaType(value string) bool {
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		return false
	}
	_, ok := previewableMediaTypes[strings.ToLower(mediaType)]
	return ok
}

// previewableMediaTypes are served with an inline disposition. DNG, HEIC, and
// HEIF are included by product decision so iPhone photos are treated uniformly,
// even though most browsers cannot render them: the client falls back to a
// download affordance when an inline image fails to load.
var previewableMediaTypes = map[string]struct{}{
	"image/jpeg":        {},
	"image/png":         {},
	"image/webp":        {},
	"image/gif":         {},
	"image/avif":        {},
	"image/heic":        {},
	"image/heif":        {},
	"image/dng":         {},
	"image/x-adobe-dng": {},
	"application/pdf":   {},
}
