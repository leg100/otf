package token

import (
	"context"
	"fmt"
	"time"

	"github.com/leg100/otf"
)

// Token is a user API token.
type Token struct {
	id          string
	createdAt   time.Time
	token       string
	description string
	// Token belongs to a user
	userID string
}

func (t *Token) ID() string           { return t.id }
func (t *Token) Token() string        { return t.token }
func (t *Token) CreatedAt() time.Time { return t.createdAt }
func (t *Token) Description() string  { return t.description }
func (t *Token) UserID() string       { return t.userID }

type TokenCreateOptions struct {
	Description string
}

type TokenService interface {
	// CreateToken creates a user token.
	CreateToken(ctx context.Context, userID string, opts *TokenCreateOptions) (*Token, error)
	// ListTokens lists API tokens for a user
	ListTokens(ctx context.Context, userID string) ([]*Token, error)
	// DeleteToken deletes a user token.
	DeleteToken(ctx context.Context, userID string, tokenID string) error
}

// TokenStore is a persistence store for user authentication tokens.
type TokenStore interface {
	// CreateToken creates a user token.
	CreateToken(ctx context.Context, token *Token) error
	// ListTokens lists user tokens.
	ListTokens(ctx context.Context, userID string) ([]*Token, error)
	// DeleteToken deletes a user token.
	DeleteToken(ctx context.Context, id string) error
}

func NewToken(uid, description string) (*Token, error) {
	t, err := otf.GenerateAuthToken("user")
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	token := Token{
		id:          otf.NewID("ut"),
		createdAt:   otf.CurrentTimestamp(),
		token:       t,
		description: description,
		userID:      uid,
	}
	return &token, nil
}
