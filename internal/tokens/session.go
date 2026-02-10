package tokens

import (
	"net/http"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/helpers"
)

const (
	// session cookie stores the session token
	SessionCookie        = "session"
	defaultSessionExpiry = 24 * time.Hour
)

func (a *Service) StartSession(w http.ResponseWriter, r *http.Request, userID resource.TfeID) error {
	expiry := internal.CurrentTimestamp(nil).Add(defaultSessionExpiry)
	token, err := a.NewToken(userID, WithExpiry(expiry))
	if err != nil {
		return err
	}
	// Set cookie to expire at same time as token
	helpers.SetCookie(w, SessionCookie, string(token), internal.Ptr(expiry))
	helpers.ReturnUserOriginalPage(w, r)

	// TODO: log username instead
	a.V(2).Info("started session", "user_id", userID)

	return nil
}
