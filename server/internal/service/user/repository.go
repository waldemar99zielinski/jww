package user

import (
	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

type Repository interface {
	Create(ctx domain.AppContext, user domain.User, uow ...unit_of_work.UnitOfWork) error
	GetByID(ctx domain.AppContext, userID uuid.UUID, uow ...unit_of_work.UnitOfWork) (*domain.User, error)
	List(ctx domain.AppContext, uow ...unit_of_work.UnitOfWork) ([]domain.User, error)
	Update(ctx domain.AppContext, user domain.User, uow ...unit_of_work.UnitOfWork) error
	Delete(ctx domain.AppContext, user domain.User, uow ...unit_of_work.UnitOfWork) error
}
