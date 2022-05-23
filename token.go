package otf

import (
	"fmt"
)

// Token is a user authentication token.
type Token struct {
	id string

	Token string

	Timestamps

	Description string

	// Token belongs to a user
	UserID string
}

func (t *Token) ID() string { return t.id }

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
