// Package auth implements login/token verification. Its repository and
// collaborator interfaces are defined here, scoped to exactly what this
// service needs — deliberately not shared with package user's own
// (larger) UserRepository interface, even though one concrete Postgres
// gateway satisfies both. See internal/service/user for the other side of
// that split.
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

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, passwordHash string) bool
}

type TokenService interface {
	Issue(userID uuid.UUID) (string, error)
	// Verify returns the underlying error for any invalid, expired, or
	// malformed token, so an implementation can log or inspect the
	// specific reason — but Service.Authenticate below deliberately
	// discards it and always surfaces a generic AuthenticationError; the
	// specific "bad token" reason never reaches callers of this service.
	Verify(token string) (*uuid.UUID, error)
}

type Service struct {
	repository     UserRepository
	passwordHasher PasswordHasher
	tokenService   TokenService
}

func NewService(repository UserRepository, passwordHasher PasswordHasher, tokenService TokenService) *Service {
	return &Service{repository: repository, passwordHasher: passwordHasher, tokenService: tokenService}
}

func (s *Service) Login(ctx domain.AppContext, username, password string) (string, error) {
	user, err := s.repository.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if user == nil || !s.passwordHasher.Verify(password, user.PasswordHash) {
		return "", domain.NewAuthenticationError()
	}
	return s.tokenService.Issue(user.ID)
}

func (s *Service) Authenticate(ctx domain.AppContext, token string) (*domain.User, error) {
	userID, err := s.tokenService.Verify(token)
	if err != nil || userID == nil {
		return nil, domain.NewAuthenticationError()
	}
	user, err := s.repository.GetByID(ctx, *userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.NewAuthenticationError()
	}
	return user, nil
}
