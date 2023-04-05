package auth

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func NewTestToken(t *testing.T, org string) *Token {
	token, _, err := NewToken(NewTokenOptions{
		Username: uuid.NewString(),
		TokenCreateOptions: TokenCreateOptions{
			Description: "lorem ipsum...",
		},
	})
	require.NoError(t, err)
	return token
}
