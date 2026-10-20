// Package webauthn_credential implements CRUD use cases for the biometric
// (WebAuthn/Face ID/Touch ID/...) credentials attached to an existing user.
// It never creates a user — see internal/service/user for that — it only
// attaches/detaches an additional login method to one that already exists.
package webauthn_credential

import (
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

type Service struct {
	repository Repository
	uowManager unit_of_work.Manager
}

func NewService(repository Repository, uowManager unit_of_work.Manager) *Service {
	return &Service{repository: repository, uowManager: uowManager}
}

func (s *Service) Create(
	ctx domain.AppContext,
	userID uuid.UUID,
	name string,
	credentialID, publicKey []byte,
	attestationType string,
	aaguid []byte,
	signCount uint32,
	transports []string,
	backupEligible bool,
	uow ...unit_of_work.UnitOfWork,
) (*domain.WebAuthnCredential, error) {
	credential := domain.WebAuthnCredential{
		ID:              uuid.New(),
		UserID:          userID,
		Name:            name,
		CredentialID:    credentialID,
		PublicKey:       publicKey,
		AttestationType: attestationType,
		AAGUID:          aaguid,
		SignCount:       signCount,
		Transports:      transports,
		BackupEligible:  backupEligible,
		CreatedAt:       time.Now().UTC(),
	}
	if err := credential.Validate(); err != nil {
		return nil, err
	}
	if err := s.repository.Create(ctx, credential, uow...); err != nil {
		return nil, err
	}
	return &credential, nil
}

// List returns every credential enrolled for userID, for the Settings
// management UI.
func (s *Service) List(ctx domain.AppContext, userID uuid.UUID, uow ...unit_of_work.UnitOfWork) ([]domain.WebAuthnCredential, error) {
	return s.repository.GetByUserID(ctx, userID, uow...)
}

// GetByCredentialID resolves which stored credential a login assertion
// claims to be. A miss returns (nil, nil) deliberately — the login workflow
// collapses it into the same generic AuthenticationError as every other
// login failure, never a typed NotFoundError, so a caller can't use this to
// enumerate which credential IDs exist.
func (s *Service) GetByCredentialID(ctx domain.AppContext, credentialID []byte, uow ...unit_of_work.UnitOfWork) (*domain.WebAuthnCredential, error) {
	return s.repository.GetByCredentialID(ctx, credentialID, uow...)
}

// TouchSignCount persists the authenticator's updated signature counter and
// last-used timestamp after a successful login assertion.
func (s *Service) TouchSignCount(ctx domain.AppContext, id uuid.UUID, signCount uint32, lastUsedAt time.Time, uow ...unit_of_work.UnitOfWork) error {
	return s.repository.UpdateSignCount(ctx, id, signCount, lastUsedAt, uow...)
}

// Delete removes a credential, but only if it belongs to userID — a
// not-found and a not-yours both 404 identically, so a caller can't use this
// to probe which credential IDs belong to someone else.
func (s *Service) Delete(ctx domain.AppContext, userID uuid.UUID, credentialID uuid.UUID) error {
	uow, err := s.uowManager.Create(ctx)
	if err != nil {
		return err
	}
	existing, err := s.repository.GetByID(ctx, credentialID, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return err
	}
	if existing == nil || existing.UserID != userID {
		_ = uow.Restore(ctx)
		return domain.NewWebAuthnCredentialNotFoundError(credentialID)
	}
	if err := s.repository.Delete(ctx, credentialID, uow); err != nil {
		_ = uow.Restore(ctx)
		return err
	}
	return uow.Finish(ctx)
}
