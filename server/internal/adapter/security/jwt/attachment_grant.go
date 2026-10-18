package jwt

import (
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"jww/internal/service/transaction_attachment"
)

const (
	attachmentGrantAudience = "jww-attachment-content"
)

type attachmentGrantClaims struct {
	AttachmentID   string `json:"attachment_id"`
	TransactionID  *int64 `json:"transaction_id,omitempty"`
	StorageVersion string `json:"storage_version"`
	Purpose        string `json:"purpose"`
	jwtlib.RegisteredClaims
}

// AttachmentGrantService issues short-lived, purpose-bound attachment
// capabilities. Its audience cannot be accepted as a user access token.
type AttachmentGrantService struct {
	codec codec
	ttl   time.Duration
}

func NewAttachmentGrantService(cfg Config, ttl time.Duration) *AttachmentGrantService {
	return &AttachmentGrantService{codec: newCodec(cfg), ttl: ttl}
}

func (s *AttachmentGrantService) Issue(claims transaction_attachment.GrantClaims) (string, error) {
	now := time.Now().UTC()
	return s.codec.sign(attachmentGrantClaims{
		AttachmentID:   claims.AttachmentID.String(),
		TransactionID:  claims.TransactionID,
		StorageVersion: claims.StorageVersion,
		Purpose:        string(claims.Purpose),
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    tokenIssuer,
			Subject:   claims.AttachmentID.String(),
			Audience:  jwtlib.ClaimStrings{attachmentGrantAudience},
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(s.ttl)),
		},
	})
}

func (s *AttachmentGrantService) Verify(tokenString string) (transaction_attachment.GrantClaims, error) {
	claims := &attachmentGrantClaims{}
	token, err := s.codec.parse(tokenString, claims,
		jwtlib.WithAudience(attachmentGrantAudience),
		jwtlib.WithIssuer(tokenIssuer),
		jwtlib.WithExpirationRequired(),
		jwtlib.WithIssuedAt(),
	)
	if err != nil {
		return transaction_attachment.GrantClaims{}, err
	}
	if !token.Valid {
		return transaction_attachment.GrantClaims{}, fmt.Errorf("security/jwt: attachment grant is not valid")
	}
	attachmentID, err := uuid.Parse(claims.AttachmentID)
	if err != nil || attachmentID == uuid.Nil || claims.Subject != claims.AttachmentID {
		return transaction_attachment.GrantClaims{}, fmt.Errorf("security/jwt: invalid attachment grant subject")
	}
	if claims.TransactionID != nil && *claims.TransactionID <= 0 {
		return transaction_attachment.GrantClaims{}, fmt.Errorf("security/jwt: invalid attachment grant transaction")
	}
	purpose, err := transaction_attachment.ParseGrantPurpose(claims.Purpose)
	if err != nil {
		return transaction_attachment.GrantClaims{}, err
	}
	return transaction_attachment.GrantClaims{
		AttachmentID:   attachmentID,
		TransactionID:  claims.TransactionID,
		StorageVersion: claims.StorageVersion,
		Purpose:        purpose,
	}, nil
}

var _ transaction_attachment.GrantTokenService = (*AttachmentGrantService)(nil)
