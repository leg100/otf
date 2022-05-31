package html

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/leg100/otf"
)

// unexported key type prevents collisions
type ctxKey int

const (
	sessionCookie = "session"

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

		cookie, err := r.Cookie(sessionCookie)
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
			setCookie(w, sessionCookie, user.Session.Token, &user.Session.Expiry)
		}

		ctx := context.WithValue(r.Context(), userCtxKey, user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// Destroy deletes the current session.
func (s *sessions) Destroy(ctx context.Context, w http.ResponseWriter) error {
	user := getUserFromContext(ctx)
	if err := s.ActiveUserService.DeleteSession(ctx, user.Session.Token); err != nil {
		return err
	}
	setCookie(w, sessionCookie, user.Session.Token, &time.Time{})

	return nil
}

func (s *sessions) IsAuthenticated(ctx context.Context) bool {
	return getUserFromContext(ctx).IsAuthenticated()
}

func getUserFromContext(ctx context.Context) *ActiveUser {
	c, ok := ctx.Value(userCtxKey).(*ActiveUser)
	if !ok {
		panic("no user in context")
	}
	return c
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
