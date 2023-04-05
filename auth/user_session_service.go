package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type (
	StartUserSessionOptions struct {
		Username *string
		Expiry   *time.Time
	}
)

func (a *service) StartSession(w http.ResponseWriter, r *http.Request, opts StartUserSessionOptions) error {
	if opts.Username == nil {
		return fmt.Errorf("missing username")
	}
	expiry := otf.CurrentTimestamp().Add(defaultExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry
	}

	token, err := jwt.NewBuilder().
		Subject(*opts.Username).
		Claim("kind", userSessionKind).
		IssuedAt(time.Now()).
		Expiration(expiry).
		Build()
	if err != nil {
		return err
	}
	serialized, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, a.key))
	if err != nil {
		return err
	}

	// Set cookie to expire at same time as token
	html.SetCookie(w, sessionCookie, string(serialized), otf.Time(expiry))
	html.ReturnUserOriginalPage(w, r)

	a.V(2).Info("started session", "username", *opts.Username)

	return nil
}
