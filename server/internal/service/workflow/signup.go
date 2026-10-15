package workflow

import (
	"jww/internal/domain"
	"jww/internal/service/user"
)

type SignupWorkflow struct {
	userService *user.Service
}

func NewSignupWorkflow(userService *user.Service) *SignupWorkflow {
	return &SignupWorkflow{userService: userService}
}

func (w *SignupWorkflow) Signup(ctx domain.AppContext, username, password, displayName string) (*domain.User, error) {
	return w.userService.Create(ctx, username, password, displayName)
}
