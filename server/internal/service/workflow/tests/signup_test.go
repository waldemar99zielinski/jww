package workflow_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/service/user"
	"jww/internal/service/workflow"
)

func TestSignupWorkflow_Signup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := newFakeUserRepo()
		uowManager := &fakeUoWManager{}
		userService := user.NewService(userRepo, fakePasswordHasher{}, uowManager)
		w := workflow.NewSignupWorkflow(userService)

		created, err := w.Signup(newAppContext(), "julcia", "s3cret", "Julcia")

		require.NoError(t, err)
		assert.Equal(t, "julcia", created.Username)
		assert.Equal(t, "Julcia", created.DisplayName)
		assert.Equal(t, "hashed:s3cret", created.PasswordHash)
	})

	t.Run("validation failure never reaches the repository", func(t *testing.T) {
		userRepo := newFakeUserRepo()
		uowManager := &fakeUoWManager{}
		userService := user.NewService(userRepo, fakePasswordHasher{}, uowManager)
		w := workflow.NewSignupWorkflow(userService)

		_, err := w.Signup(newAppContext(), "", "s3cret", "Julcia")

		require.Error(t, err)
	})
}
