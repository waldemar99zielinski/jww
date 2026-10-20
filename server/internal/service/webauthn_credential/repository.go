package webauthn_credential

import (
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

type Repository interface {
	Create(ctx domain.AppContext, credential domain.WebAuthnCredential, uow ...unit_of_work.UnitOfWork) error
	GetByID(ctx domain.AppContext, id uuid.UUID, uow ...unit_of_work.UnitOfWork) (*domain.WebAuthnCredential, error)
	GetByCredentialID(ctx domain.AppContext, credentialID []byte, uow ...unit_of_work.UnitOfWork) (*domain.WebAuthnCredential, error)
	GetByUserID(ctx domain.AppContext, userID uuid.UUID, uow ...unit_of_work.UnitOfWork) ([]domain.WebAuthnCredential, error)
	UpdateSignCount(ctx domain.AppContext, id uuid.UUID, signCount uint32, lastUsedAt time.Time, uow ...unit_of_work.UnitOfWork) error
	Delete(ctx domain.AppContext, id uuid.UUID, uow ...unit_of_work.UnitOfWork) error
}
