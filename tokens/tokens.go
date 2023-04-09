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

	userSessionKind kind = "user_session"
	runTokenKind    kind = "run_token"
	agentTokenKind  kind = "agent_token"
	userTokenKind   kind = "user_token"
)

type (
	// the kind of authentication token: user session, user token, agent token, etc
	kind string

	newTokenOptions struct {
		key     jwk.Key
		kind    kind
		subject string
		expiry  *time.Time
		claims  map[string]string
	}
)

func newToken(opts newTokenOptions) ([]byte, error) {
	builder := jwt.NewBuilder().
		Subject(opts.subject).
		Claim("kind", opts.kind).
		IssuedAt(time.Now())
	for k, v := range opts.claims {
		builder = builder.Claim(k, v)
	}
	if opts.expiry != nil {
		builder = builder.Expiration(*opts.expiry)
	}
	token, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return jwt.Sign(token, jwt.WithKey(jwa.HS256, opts.key))
}
