package tokens

import (
	"fmt"
	"net/http"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

const (
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
)

func NewSessionToken(key jwk.Key, username string, expiry time.Time) (string, error) {
	token, err := NewToken(NewTokenOptions{
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
	token, err := NewSessionToken(a.key, *opts.Username, expiry)
	if err != nil {
		return err
	}
	// Set cookie to expire at same time as token
	html.SetCookie(w, sessionCookie, string(token), internal.Time(expiry))
	html.ReturnUserOriginalPage(w, r)

	a.V(2).Info("started session", "username", *opts.Username)

	return nil
}
