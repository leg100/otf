package tokens

import (
	"fmt"
	"net/http"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
)

type (
	StartSessionOptions struct {
		Username *string
		Expiry   *time.Time
	}

	sessionService interface {
		StartSession(w http.ResponseWriter, r *http.Request, opts StartSessionOptions) error
	}
)

func (a *service) StartSession(w http.ResponseWriter, r *http.Request, opts StartSessionOptions) error {
	if opts.Username == nil {
		return fmt.Errorf("missing username")
	}
	expiry := otf.CurrentTimestamp().Add(defaultSessionExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry
	}
	token, err := newToken(newTokenOptions{
		key:     a.key,
		subject: *opts.Username,
		kind:    userSessionKind,
		expiry:  &expiry,
	})
	if err != nil {
		return err
	}
	// Set cookie to expire at same time as token
	html.SetCookie(w, sessionCookie, string(token), otf.Time(expiry))
	html.ReturnUserOriginalPage(w, r)

	a.V(2).Info("started session", "username", *opts.Username)

	return nil
}
