package otf

import (
	"context"
	"fmt"
	"time"
)

// Token is a user authentication token.
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

// TokenStore is a persistence store for user authentication tokens.
type TokenStore interface {
	// CreateToken creates a user token.
	CreateToken(ctx context.Context, token *Token) error
	// DeleteToken deletes a user token.
	DeleteToken(ctx context.Context, id string) error
}

func NewToken(uid, description string) (*Token, error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	session := Token{
		id:          NewID("ut"),
		createdAt:   CurrentTimestamp(),
		token:       token,
		description: description,
		userID:      uid,
	}
	return &session, nil
}
