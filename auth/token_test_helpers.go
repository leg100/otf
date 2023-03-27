package auth

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func NewTestToken(t *testing.T, org string) *Token {
	token, err := NewToken(uuid.NewString(), "lorem ipsum...")
	require.NoError(t, err)
	return token
}
