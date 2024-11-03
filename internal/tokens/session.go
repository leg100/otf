package tokens

import (
	"net/http"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/resource"
)

const (
	// session cookie stores the session token
	SessionCookie        = "session"
	defaultSessionExpiry = 24 * time.Hour
)

type (
	StartSessionOptions struct {
		UserID resource.ID
		Expiry *time.Time
	}

	// sessionFactory constructs new sessions.
	sessionFactory struct {
		*factory
	}
)

func (f *sessionFactory) NewSessionToken(userID resource.ID, expiry time.Time) (string, error) {
	token, err := f.NewToken(NewTokenOptions{
		ID:     userID,
		Expiry: &expiry,
	})
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func (a *Service) StartSession(w http.ResponseWriter, r *http.Request, opts StartSessionOptions) error {
	expiry := internal.CurrentTimestamp(nil).Add(defaultSessionExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry
	}
	token, err := a.NewSessionToken(opts.UserID, expiry)
	if err != nil {
		return err
	}
	// Set cookie to expire at same time as token
	html.SetCookie(w, SessionCookie, string(token), internal.Time(expiry))
	html.ReturnUserOriginalPage(w, r)

	// TODO: log username instead
	a.V(2).Info("started session", "user_id", opts.UserID)

	return nil
}
