// Package user implements user CRUD. Its UserRepository is deliberately
// its own interface (Create/GetByID/List/Update/Delete), separate from
// package auth's smaller UserRepository (GetByID/GetByUsername) — the two
// are independently owned even though one concrete Postgres gateway
// satisfies both.
package user

import (
	"time"

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

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, passwordHash string) bool
}

type Service struct {
	repository     Repository
	passwordHasher PasswordHasher
	uowManager     unit_of_work.Manager
}

func NewService(repository Repository, passwordHasher PasswordHasher, uowManager unit_of_work.Manager) *Service {
	return &Service{repository: repository, passwordHasher: passwordHasher, uowManager: uowManager}
}

func (s *Service) Create(ctx domain.AppContext, username, password, displayName string, uow ...unit_of_work.UnitOfWork) (*domain.User, error) {
	passwordHash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return nil, err
	}
	user := domain.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		CreatedAt:    time.Now().UTC(),
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	if err := s.repository.Create(ctx, user, uow...); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) GetByID(ctx domain.AppContext, userID uuid.UUID, uow ...unit_of_work.UnitOfWork) (*domain.User, error) {
	user, err := s.repository.GetByID(ctx, userID, uow...)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.NewUserNotFoundError(userID)
	}
	return user, nil
}

func (s *Service) List(ctx domain.AppContext, uow ...unit_of_work.UnitOfWork) ([]domain.User, error) {
	return s.repository.List(ctx, uow...)
}

func (s *Service) Update(ctx domain.AppContext, user domain.User) (*domain.User, error) {
	if err := user.Validate(); err != nil {
		return nil, err
	}
	active, err := s.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	existing, err := s.repository.GetByID(ctx, user.ID, active)
	if err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	if existing == nil {
		_ = active.Restore(ctx)
		return nil, domain.NewUserNotFoundError(user.ID)
	}
	if err := s.repository.Update(ctx, user, active); err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	if err := active.Finish(ctx); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) Delete(ctx domain.AppContext, user domain.User) error {
	active, err := s.uowManager.Create(ctx)
	if err != nil {
		return err
	}
	existing, err := s.repository.GetByID(ctx, user.ID, active)
	if err != nil {
		_ = active.Restore(ctx)
		return err
	}
	if existing == nil {
		_ = active.Restore(ctx)
		return domain.NewUserNotFoundError(user.ID)
	}
	if err := s.repository.Delete(ctx, user, active); err != nil {
		_ = active.Restore(ctx)
		return err
	}
	return active.Finish(ctx)
}
