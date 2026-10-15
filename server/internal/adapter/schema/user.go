package schema

import (
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
)

type UserCreateRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

func (r UserCreateRequest) Validate() error {
	return firstError(
		minLength("username", r.Username, 1),
		minLength("password", r.Password, 8),
		minLength("display_name", r.DisplayName, 1),
	)
}

type UserUpdateRequest struct {
	DisplayName string `json:"display_name"`
}

func (r UserUpdateRequest) Validate() error {
	return minLength("display_name", r.DisplayName, 1)
}

// UserResponse deliberately omits PasswordHash — never serialize it back to
// a client.
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

func UserToResponse(u domain.User) UserResponse {
	return UserResponse{ID: u.ID, Username: u.Username, DisplayName: u.DisplayName, CreatedAt: u.CreatedAt}
}
