// Package session handles user sessions.
package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/path"
)

const (
	// session cookie stores the session token
	SessionCookie        = "session"
	defaultSessionExpiry = 24 * time.Hour
)

// Authenticator authenticates requests to the UI.
type Authenticator struct {
	Client AuthenticatorClient
}

type AuthenticatorClient interface {
	GetSubject(ctx context.Context, token []byte) (authz.Subject, error)
}

func (a *Authenticator) Authenticate(w http.ResponseWriter, r *http.Request) (authz.Subject, error) {
	if !strings.HasPrefix(r.URL.Path, path.Prefix) {
		// This is a non-UI request; session authenticator only authenticates
		// access to the UI.
		return nil, nil
	}
	// Any error results in the user being redirected to the login page with a
	// flash message explaining why.
	subj, err := a.authWithError(r)
	if err != nil {
		helpers.FlashError(w, err.Error())
		helpers.SendUserToLoginPage(w, r)
		return nil, nil
	}
	// Successfully authenticated.
	return subj, nil
}

func (a *Authenticator) authWithError(r *http.Request) (authz.Subject, error) {
	cookie, err := r.Cookie(SessionCookie)
	if err == http.ErrNoCookie {
		return nil, fmt.Errorf("you need to login to access the requested page")
	}
	user, err := a.Client.GetSubject(r.Context(), []byte(cookie.Value))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired()) {
			return nil, fmt.Errorf("session expired")
		}
		return nil, fmt.Errorf("unable to verify session token: %w", err)
	}
	return user, nil
}
