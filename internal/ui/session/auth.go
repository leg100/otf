// Package session handles user sessions.
package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/leg100/otf/internal/authz"
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
	cookie, err := r.Cookie(SessionCookie)
	if err == http.ErrNoCookie {
		return nil, nil
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
