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
	NewTokenOption func(*jwt.Builder) *jwt.Builder

	// tokenFactory constructs new tokens using a JWK
	tokenFactory struct {
		symKey     jwk.Key
		PrivateKey jwk.Key
	}
)

func WithExpiry(exp time.Time) NewTokenOption {
	return func(builder *jwt.Builder) *jwt.Builder {
		return builder.Expiration(exp)
	}
}

func (f *tokenFactory) NewToken(subjectID resource.TfeID, opts ...NewTokenOption) ([]byte, error) {
	builder := jwt.NewBuilder().
		Subject(subjectID.String()).
		IssuedAt(time.Now())
	//for k, v := range opts.Claims {
	//	builder = builder.Claim(k, v)
	//}
	for _, fn := range opts {
		builder = fn(builder)
	}
	token, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return jwt.Sign(token, jwt.WithKey(jwa.HS256, f.symKey))
}
