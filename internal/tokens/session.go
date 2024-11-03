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

// sessionFactory constructs new sessions.
type sessionFactory struct {
	*tokenFactory
}

func (a *Service) StartSession(w http.ResponseWriter, r *http.Request, userID resource.ID) error {
	expiry := internal.CurrentTimestamp(nil).Add(defaultSessionExpiry)
	token, err := a.NewToken(userID, NewTokenOptions{Expiry: &expiry})
	if err != nil {
		return err
	}
	// Set cookie to expire at same time as token
	html.SetCookie(w, SessionCookie, string(token), internal.Time(expiry))
	html.ReturnUserOriginalPage(w, r)

	// TODO: log username instead
	a.V(2).Info("started session", "user_id", userID)

	return nil
}
