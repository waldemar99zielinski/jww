package auth

import (
	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

type UserRepository interface {
	GetByID(ctx domain.AppContext, userID uuid.UUID, uow ...unit_of_work.UnitOfWork) (*domain.User, error)
	GetByUsername(ctx domain.AppContext, username string, uow ...unit_of_work.UnitOfWork) (*domain.User, error)
}
