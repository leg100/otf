package otf

import (
	"encoding/base64"
	"fmt"
	"math/rand"
)

// Token is a user session
type Token struct {
	ID string `db:"token_id"`

	Token string

	Timestamps

	Description string

	// Token belongs to a user
	UserID string
}

func NewToken(uid, description string) (*Token, error) {
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	session := Token{
		ID:          NewID("ut"),
		Token:       token,
		Timestamps:  NewTimestamps(),
		Description: description,
		UserID:      uid,
	}

	return &session, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
