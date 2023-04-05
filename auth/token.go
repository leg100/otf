package auth

import (
	"time"

	"github.com/leg100/otf"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type (
	// Token is a user API token.
	Token struct {
		ID          string
		CreatedAt   time.Time
		Description string
		Username    string // Token belongs to a user
	}

	TokenCreateOptions struct {
		Description string
	}

	NewTokenOptions struct {
		TokenCreateOptions
		Username string
		key      jwk.Key
	}
)

func NewToken(opts NewTokenOptions) (*Token, []byte, error) {
	token, err := jwt.NewBuilder().
		Claim("kind", registrySessionKind).
		IssuedAt(time.Now()).
		Build()
	if err != nil {
		return nil, nil, err
	}
	serialized, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, opts.key))
	if err != nil {
		return nil, nil, err
	}
	ut := Token{
		ID:          otf.NewID("ut"),
		CreatedAt:   otf.CurrentTimestamp(),
		Description: opts.Description,
		Username:    opts.Username,
	}
	return &ut, serialized, nil
}
