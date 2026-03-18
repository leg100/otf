// Package tokens manages token authentication
package tokens

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/leg100/otf/internal/resource"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// factory constructs new tokens using a JWK
type factory struct {
	key jwk.Key
}

func (f *factory) NewToken(subjectID resource.TfeID, expiry *time.Time) ([]byte, error) {
	builder := jwt.NewBuilder().
		Subject(subjectID.String()).
		IssuedAt(time.Now())
	if expiry != nil {
		builder = builder.Expiration(*expiry)
	}
	// TODO: permit caller to add claims
	//
	//for k, v := range opts.Claims {
	//	builder = builder.Claim(k, v)
	//}
	token, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return jwt.Sign(token, jwt.WithKey(jwa.HS256, f.key))
}

func ParseBearerToken(r *http.Request) (string, error) {
	bearer := r.Header.Get("Authorization")
	if bearer == "" {
		return "", nil
	}
	splitToken := strings.Split(bearer, "Bearer ")
	if len(splitToken) != 2 {
		return "", fmt.Errorf("malformed bearer token")
	}
	return splitToken[1], nil
}
