package html

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/leg100/otf"
)

// unexported key type prevents collisions
type ctxKey int

const (
	userCtxKey = iota
)

// sessions is a user session manager.
type sessions struct {
	// perform actions against system users and sessions
	*ActiveUserService

	// session cookie config
	cookie *http.Cookie
}

// Load provides middleware that loads and attaches the User to the current
// request's context. It looks for a cookie containing a token on the request
// and if found uses it retrieve and attach the User. Otherwise a new session is
// created for the anonymous user.
func (s *sessions) Load(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// the user to attach to the request ctx
		var user *ActiveUser

		cookie, err := r.Cookie(s.cookie.Name)
		if err == nil {
			user, err = s.ActiveUserService.Get(r.Context(), cookie.Value)
			if err != otf.ErrResourceNotFound && err != nil {
				// encountered error other than not found error
				writeError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// cookie nor user found
		if err != nil {
			user, err = s.ActiveUserService.NewAnonymousSession(r)
			if err != nil {
				writeError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		ctx := context.WithValue(r.Context(), userCtxKey, user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// TransferSession transfers the active session to the given user. Note: after
// this returns, the context will refer to a user without an active session!
func (s *sessions) TransferSession(ctx context.Context, to *otf.User) error {
	existing := s.getUserFromContext(ctx)

	existing.TransferSession(existing.Session, to)

	if err := s.UserService.UpdateSession(ctx, to, existing.Session); err != nil {
		return err
	}

	return nil
}

// Destroy deletes the current session.
func (s *sessions) Destroy(ctx context.Context) error {
	user := s.getUserFromContext(ctx)
	return s.ActiveUserService.DeleteSession(ctx, user.Session.Token)
}

func (s *sessions) IsAuthenticated(ctx context.Context) bool {
	return s.getUserFromContext(ctx).IsAuthenticated()
}

func (s *sessions) getUserFromContext(ctx context.Context) *ActiveUser {
	c, ok := ctx.Value(userCtxKey).(*ActiveUser)
	if !ok {
		panic("no user in context")
	}
	return c
}

// PopFlash retrieves a flash message from the current session. The message is
// thereafter discarded. Nil is returned if there is no flash message.
func (s *sessions) PopFlash(r *http.Request) (*otf.Flash, error) {
	user := s.getUserFromContext(r.Context())

	flash := user.PopFlash()

	if flash == nil {
		return nil, nil
	}

	// Discard flash in store
	if err := s.UserService.UpdateSession(r.Context(), user.User, user.Session); err != nil {
		return nil, fmt.Errorf("saving flash message in session backend: %w", err)
	}

	return flash, nil
}

func (s *sessions) FlashSuccess(r *http.Request, msg ...string) error {
	return s.flash(r, otf.FlashSuccessType, msg...)
}

func (s *sessions) FlashError(r *http.Request, msg ...string) error {
	return s.flash(r, otf.FlashErrorType, msg...)
}

func (s *sessions) flash(r *http.Request, t otf.FlashType, msg ...string) error {
	user := s.getUserFromContext(r.Context())

	user.SetFlash(t, msg)

	if err := s.UserService.UpdateSession(r.Context(), user.User, user.Session); err != nil {
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
