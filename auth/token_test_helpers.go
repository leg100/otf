package auth

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func NewTestToken(t *testing.T, org string) *Token {
	token, _, err := NewToken(NewTokenOptions{
		CreateTokenOptions: CreateTokenOptions{
			Description: "lorem ipsum...",
		},
		Username: uuid.NewString(),
		key: newTestJWK(t, "something_secret"),
	})
	require.NoError(t, err)
	return token
}
