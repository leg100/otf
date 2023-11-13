package tokens

import (
	"fmt"
	"net/http"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
)

const (
	// session cookie stores the session token
	SessionCookie             = "session"
	userSessionKind      Kind = "user_session"
	defaultSessionExpiry      = 24 * time.Hour
)

type (
	StartSessionOptions struct {
		Username *string
		Expiry   *time.Time
	}

	sessionService interface {
		StartSession(w http.ResponseWriter, r *http.Request, opts StartSessionOptions) error
	}

	// sessionFactory constructs new sessions.
	sessionFactory struct {
		*factory
	}
)

func (f *sessionFactory) NewSessionToken(username string, expiry time.Time) (string, error) {
	token, err := f.NewToken(NewTokenOptions{
		Subject: username,
		Kind:    userSessionKind,
		Expiry:  &expiry,
	})
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func (a *service) StartSession(w http.ResponseWriter, r *http.Request, opts StartSessionOptions) error {
	if opts.Username == nil {
		return fmt.Errorf("missing username")
	}
	expiry := internal.CurrentTimestamp(nil).Add(defaultSessionExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry
	}
	token, err := a.NewSessionToken(*opts.Username, expiry)
	if err != nil {
		return err
	}
	// Set cookie to expire at same time as token
	html.SetCookie(w, SessionCookie, string(token), internal.Time(expiry))
	html.ReturnUserOriginalPage(w, r)

	a.V(2).Info("started session", "username", *opts.Username)

	return nil
}
