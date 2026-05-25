package user

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func NewTestUser(t *testing.T) *User {
	user, err := NewUser(uuid.NewString())
	require.NoError(t, err)
	return user
}
