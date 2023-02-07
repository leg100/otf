package user

import (
	"testing"

	"github.com/google/uuid"
)

func NewTestUser(t *testing.T, opts ...NewUserOption) *User {
	return NewUser(uuid.NewString(), opts...)
}
