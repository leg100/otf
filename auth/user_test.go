package auth

import (
	"testing"

	"github.com/google/uuid"
)

func NewTestUser(t *testing.T, opts ...newUserOption) *User {
	user := NewUser(uuid.NewString())
	for _, fn := range opts {
		fn(user)
	}
	return user
}
