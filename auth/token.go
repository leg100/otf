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

	// CreateTokenOptions are options for creating a user token via the service
	// endpoint
	CreateTokenOptions struct {
		Description string
	}

	// NewTokenOptions are options for constructing a user token via the
	// constructor.
	NewTokenOptions struct {
		CreateTokenOptions
		Username string
		key      jwk.Key
	}
)

func NewToken(opts NewTokenOptions) (*Token, []byte, error) {
	ut := Token{
		ID:          otf.NewID("ut"),
		CreatedAt:   otf.CurrentTimestamp(),
		Description: opts.Description,
		Username:    opts.Username,
	}
	token, err := jwt.NewBuilder().
		Subject(ut.ID).
		Claim("kind", userTokenKind).
		IssuedAt(time.Now()).
		Build()
	if err != nil {
		return nil, nil, err
	}
	serialized, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, opts.key))
	if err != nil {
		return nil, nil, err
	}
	return &ut, serialized, nil
}
