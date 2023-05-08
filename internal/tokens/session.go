package tokens

import (
	"fmt"
	"net/http"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type (
	StartSessionOptions struct {
		Username *string
		Expiry   *time.Time
	}

	sessionService interface {
		StartSession(w http.ResponseWriter, r *http.Request, opts StartSessionOptions) error
	}

	NewSessionTokenOptions struct {
		Key      jwk.Key
		Username string
		Expiry   *time.Time
	}
)

func (a *service) StartSession(w http.ResponseWriter, r *http.Request, opts StartSessionOptions) error {
	if opts.Username == nil {
		return fmt.Errorf("missing username")
	}
	expiry := internal.CurrentTimestamp().Add(defaultSessionExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry
	}
	token, err := NewSessionToken(NewSessionTokenOptions{
		Key:      a.key,
		Username: *opts.Username,
		Expiry:   &expiry,
	})
	if err != nil {
		return err
	}
	// Set cookie to expire at same time as token
	html.SetCookie(w, sessionCookie, string(token), internal.Time(expiry))
	html.ReturnUserOriginalPage(w, r)

	a.V(2).Info("started session", "username", *opts.Username)

	return nil
}

func NewSessionToken(opts NewSessionTokenOptions) ([]byte, error) {
	expiry := internal.CurrentTimestamp().Add(defaultSessionExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry
	}
	return newToken(newTokenOptions{
		key:     opts.Key,
		subject: opts.Username,
		kind:    userSessionKind,
		expiry:  &expiry,
	})
}
