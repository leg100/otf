package html

import (
	"context"
	"fmt"
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
	otf.UserService

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
		var user *otf.User

		cookie, err := r.Cookie(s.cookie.Name)
		if err == nil {
			user, err = s.UserService.Get(r.Context(), otf.UserSpecifier{Token: &cookie.Value})
			if err != otf.ErrResourceNotFound && err != nil {
				// encountered error other than not found error
				writeError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// cookie nor user found
		if err != nil {
			user, err = s.UserService.NewAnonymousSession(r.Context())
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

func (s *sessions) Put(ctx context.Context) {
	user := s.getUserFromContext(ctx)
}

func (s *sessions) getUserFromContext(ctx context.Context) *otf.User {
	c, ok := ctx.Value(userCtxKey).(*otf.User)
	if !ok {
		panic("no user in context")
	}
	return c
}

// PopFlashMessages retrieves all flash messages from the current session. The
// messages are thereafter discarded.
func (s *sessions) PopAllFlash(r *http.Request) (msgs []Flash) {
	if msg := s.Pop(r.Context(), otf.FlashSessionKey); msg != nil {
		msgs = append(msgs, msg.(Flash))
	}
	return
}

func (s *sessions) FlashSuccess(r *http.Request, msg ...string) error {
	return s.flash(r, FlashSuccessType, msg...)
}

func (s *sessions) FlashError(r *http.Request, msg ...string) error {
	return s.flash(r, FlashErrorType, msg...)
}

func (s *sessions) flash(r *http.Request, t FlashType, msg ...string) error {
	user := s.getUserFromContext(r.Context())
	if user.ActiveSession == nil {
		return fmt.Errorf("user %s has no active session", user.ID)
	}

	user.ActiveSession.Data[otf.FlashSessionKey] = Flash{
		Type:    t,
		Message: fmt.Sprint(convertStringSliceToInterfaceSlice(msg)...),
	}

	if err := s.UserService.UpdateSession(r.Context(), user.ActiveSession); err != nil {
		return fmt.Errorf("saving flash message in session backend: %w", err)
	}

	return nil
}

func (s *sessions) CurrentUser(r *http.Request) string {
	return s.GetString(r.Context(), otf.UsernameSessionKey)
}

type Flash struct {
	Type    FlashType
	Message string
}

type FlashType string

const (
	FlashSuccessType = "success"
	FlashErrorType   = "error"
)

func convertStringSliceToInterfaceSlice(ss []string) (is []interface{}) {
	for _, s := range ss {
		is = append(is, interface{}(s))
	}
	return
}
