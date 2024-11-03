// Package tokens manages token authentication
package tokens

import (
	"time"

	"github.com/leg100/otf/internal/resource"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type (
	NewTokenOptions struct {
		ID     resource.ID
		Expiry *time.Time
		Claims map[string]string
	}

	// factory constructs new tokens using a jwk
	factory struct {
		key jwk.Key
	}
)

func (f *factory) NewToken(opts NewTokenOptions) ([]byte, error) {
	builder := jwt.NewBuilder().
		Subject(opts.ID.String()).
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
	return jwt.Sign(token, jwt.WithKey(jwa.HS256, f.key))
}
