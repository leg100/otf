package html

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/leg100/otf"
)

// unexported key type prevents collisions
type ctxKey int

const (
	sessionCookieName = "session"

	userCtxKey ctxKey = iota
)

// sessions is a user session manager.
type sessions struct {
	// perform actions against system users and sessions
	*ActiveUserService
}

// Load provides middleware that loads and attaches the User to the current
// request's context. It looks for a cookie containing a token on the request
// and if found uses it to retrieve and attach the User. Otherwise a new session
// is created for the anonymous user.
func (s *sessions) Load(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// the user to attach to the request ctx
		var user *ActiveUser

		cookie, err := r.Cookie(sessionCookieName)
		if err == nil {
			user, err = s.ActiveUserService.Get(r.Context(), cookie.Value)
			if err != otf.ErrResourceNotFound && err != nil {
				// encountered error other than not found error
				writeError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// cookie or user not found
		if err != nil {
			user, err = s.ActiveUserService.NewAnonymousSession(r)
			if err != nil {
				writeError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// set cookie on response
			setCookie(w, user.Session.Token, user.Session.Expiry)
		}

		ctx := context.WithValue(r.Context(), userCtxKey, user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func setCookie(w http.ResponseWriter, token string, expiry time.Time) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	if expiry.IsZero() {
		// Purge cookie from browser.
		cookie.Expires = time.Unix(1, 0)
		cookie.MaxAge = -1
	} else {
		// Round up to the nearest second.
		cookie.Expires = time.Unix(expiry.Unix()+1, 0)
		cookie.MaxAge = int(time.Until(expiry).Seconds() + 1)
	}

	w.Header().Add("Set-Cookie", cookie.String())
	w.Header().Add("Cache-Control", `no-cache="Set-Cookie"`)
}

// Destroy deletes the current session.
func (s *sessions) Destroy(ctx context.Context, w http.ResponseWriter) error {
	user := GetUserFromContext(ctx)
	if err := s.ActiveUserService.DeleteSession(ctx, user.Session.Token); err != nil {
		return err
	}
	setCookie(w, user.Session.Token, time.Time{})

	return nil
}

func (s *sessions) IsAuthenticated(ctx context.Context) bool {
	return GetUserFromContext(ctx).IsAuthenticated()
}

func GetUserFromContext(ctx context.Context) *ActiveUser {
	c, ok := ctx.Value(userCtxKey).(*ActiveUser)
	if !ok {
		panic("no user in context")
	}
	return c
}

func (s *sessions) GetSessionFromContext(ctx context.Context) *otf.Session {
	return GetUserFromContext(ctx).Session
}

// PopFlash retrieves a flash message from the current session. The message is
// thereafter discarded. Nil is returned if there is no flash message.
func (s *sessions) PopFlash(r *http.Request) (*otf.Flash, error) {
	session := s.GetSessionFromContext(r.Context())

	return s.UserService.PopFlash(r.Context(), session.Token)
}

func (s *sessions) FlashSuccess(r *http.Request, msg ...interface{}) error {
	return s.flash(r, otf.FlashSuccess(msg...))
}

func (s *sessions) FlashError(r *http.Request, msg ...interface{}) error {
	return s.flash(r, otf.FlashError(msg...))
}

func (s *sessions) flash(r *http.Request, flash *otf.Flash) error {
	session := s.GetSessionFromContext(r.Context())

	if err := s.UserService.SetFlash(r.Context(), session.Token, flash); err != nil {
		return fmt.Errorf("saving flash message in session backend: %w", err)
	}

	return nil
}

func newSessionData(r *http.Request) (*otf.SessionData, error) {
	addr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, err
	}

	data := otf.SessionData{
		Address: addr,
	}

	return &data, nil
}
