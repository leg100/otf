package otf

import (
	"context"
	"fmt"
)

// Token is a user authentication token.
type Token struct {
	id    string
	Token string
	Timestamps
	Description string
	// Token belongs to a user
	UserID string
}

func (t *Token) ID() string { return t.id }

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
		Token:       token,
		Description: description,
		UserID:      uid,
	}
	return &session, nil
}
