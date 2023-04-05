package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type (
	StatelessSessionService interface {
		StartSession(w http.ResponseWriter, r *http.Request, opts CreateStatelessSessionOptions) error
	}
	CreateStatelessSessionOptions struct {
		Username *string
		Expiry   *time.Time
	}
	statelessSessionService struct {
		logr.Logger
		key jwk.Key
	}
)

func newStatelessSessionService(logger logr.Logger, secret string) (*statelessSessionService, error) {
	key, err := jwk.FromRaw(secret)
	if err != nil {
		return nil, err
	}
	return &statelessSessionService{Logger: logger, key: key}, nil
}

func (a *statelessSessionService) StartSession(w http.ResponseWriter, r *http.Request, opts CreateStatelessSessionOptions) error {
	if opts.Username == nil {
		return fmt.Errorf("missing username")
	}
	expiry := otf.CurrentTimestamp().Add(defaultExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry
	}

	token, err := jwt.NewBuilder().
		Subject(*opts.Username).
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
