package currency

import (
	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

type Repository interface {
	Create(ctx domain.AppContext, currency domain.Currency, uow ...unit_of_work.UnitOfWork) error
	GetByCode(ctx domain.AppContext, code string, uow ...unit_of_work.UnitOfWork) (*domain.Currency, error)
	List(ctx domain.AppContext, uow ...unit_of_work.UnitOfWork) ([]domain.Currency, error)
	Update(ctx domain.AppContext, currency domain.Currency, uow ...unit_of_work.UnitOfWork) error
	Delete(ctx domain.AppContext, currency domain.Currency, uow ...unit_of_work.UnitOfWork) error
}
