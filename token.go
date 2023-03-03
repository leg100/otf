package otf

import (
	"context"
	"fmt"
	"time"
)

type (
	// Token is a user API token.
	Token struct {
		ID          string
		CreatedAt   time.Time
		Token       string
		Description string
		// Token belongs to a user
		UserID string
	}

	TokenCreateOptions struct {
		Description string
	}

	TokenService interface {
		// CreateToken creates a user token.
		CreateToken(ctx context.Context, userID string, opts *TokenCreateOptions) (*Token, error)
		// ListTokens lists API tokens for a user
		ListTokens(ctx context.Context, userID string) ([]*Token, error)
		// DeleteToken deletes a user token.
		DeleteToken(ctx context.Context, userID string, tokenID string) error
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
	t, err := GenerateAuthToken("user")
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	token := Token{
		ID:          NewID("ut"),
		CreatedAt:   CurrentTimestamp(),
		Token:       t,
		Description: description,
		UserID:      uid,
	}
	return &token, nil
}
