package html

import (
	"net/http"
	"time"
)

const (
	// organizationCookie stores the current organization for the session
	organizationCookie = "organization"
)

// SetCookie sets a cookie on the http response. A nil expiry sets no expiry,
// and a zero expiry sets it to be purged from the browser.
func SetCookie(w http.ResponseWriter, name, value string, expiry *time.Time) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	if expiry != nil {
		if (*expiry).IsZero() {
			// Purge cookie from browser.
			cookie.Expires = time.Unix(1, 0)
			cookie.MaxAge = -1
		} else {
			// Round up to the nearest second.
			cookie.Expires = time.Unix((*expiry).Unix()+1, 0)
			cookie.MaxAge = int(time.Until(*expiry).Seconds() + 1)
		}
	}

	w.Header().Add("Set-Cookie", cookie.String())
	w.Header().Add("Cache-Control", `no-cache="Set-Cookie"`)
}
