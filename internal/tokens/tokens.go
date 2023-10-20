// Package tokens manages token authentication
package tokens

import (
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	defaultSessionExpiry = 24 * time.Hour

	userSessionKind       Kind = "user_session"
	runTokenKind          Kind = "run_token"
	agentTokenKind        Kind = "agent_token"
	userTokenKind         Kind = "user_token"
	organizationTokenKind Kind = "organization_token"
	teamTokenKind         Kind = "team_token"
)

type (
	// the Kind of authentication token: user session, user token, agent token, etc
	Kind string

	NewTokenOptions struct {
		key     jwk.Key
		Kind    Kind
		Subject string
		Expiry  *time.Time
		Claims  map[string]string
	}
)

func NewToken(opts NewTokenOptions) ([]byte, error) {
	builder := jwt.NewBuilder().
		Subject(opts.Subject).
		Claim("kind", opts.Kind).
		IssuedAt(time.Now())
	for k, v := range opts.Claims {
		builder = builder.Claim(k, v)
	}
	if opts.Expiry != nil {
		builder = builder.Expiration(*opts.Expiry)
	}
	token, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return jwt.Sign(token, jwt.WithKey(jwa.HS256, opts.key))
}
