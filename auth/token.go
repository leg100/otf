package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/leg100/otf"
)

type (
	// Token is a user API token.
	Token struct {
		ID          string
		CreatedAt   time.Time
		Token       string
		Description string
		Username    string // Token belongs to a user
	}

	TokenCreateOptions struct {
		Description string
	}

	TokenService interface {
		// CreateToken creates a user token.
		CreateToken(ctx context.Context, username string, opts *TokenCreateOptions) (*Token, error)
		// ListTokens lists API tokens for a user
		ListTokens(ctx context.Context, username string) ([]*Token, error)
		// DeleteToken deletes a user token.
		DeleteToken(ctx context.Context, username string, tokenID string) error
	}

	// TokenStore is a persistence store for user authentication tokens.
	TokenStore interface {
		// CreateToken creates a user token.
		CreateToken(ctx context.Context, token *Token) error
		// ListTokens lists user tokens.
		ListTokens(ctx context.Context, userID string) ([]*Token, error)
		// DeleteToken deletes a user token.
		DeleteToken(ctx context.Context, id string) error
	}
)

func NewToken(uid, description string) (*Token, error) {
	t, err := otf.GenerateAuthToken("user")
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	token := Token{
		ID:          otf.NewID("ut"),
		CreatedAt:   otf.CurrentTimestamp(),
		Token:       t,
		Description: description,
		Username:    uid,
	}
	return &token, nil
}
