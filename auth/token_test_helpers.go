package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func NewTestToken(t *testing.T, org string) *Token {
	token, err := NewToken(uuid.NewString(), "lorem ipsum...")
	require.NoError(t, err)
	return token
}

type fakeTokenService struct {
	token *Token

	tokenService
}

func (f *fakeTokenService) CreateToken(context.Context, string, *TokenCreateOptions) (*Token, error) {
	return f.token, nil
}

func (f *fakeTokenService) ListTokens(context.Context, string) ([]*Token, error) {
	return []*Token{f.token}, nil
}

func (f *fakeTokenService) DeleteToken(context.Context, string, string) error {
	return nil
}
