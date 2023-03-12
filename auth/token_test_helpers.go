package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestToken(t *testing.T, org string) *otf.Token {
	token, err := otf.NewToken(uuid.NewString(), "lorem ipsum...")
	require.NoError(t, err)
	return token
}

type fakeTokenService struct {
	token *otf.Token

	tokenService
}

func (f *fakeTokenService) CreateToken(context.Context, string, *otf.TokenCreateOptions) (*otf.Token, error) {
	return f.token, nil
}

func (f *fakeTokenService) ListTokens(context.Context, string) ([]*otf.Token, error) {
	return []*otf.Token{f.token}, nil
}

func (f *fakeTokenService) DeleteToken(context.Context, string, string) error {
	return nil
}
