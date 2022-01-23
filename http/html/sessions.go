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
			user, err = s.ActiveUserService.NewAnonymousSession(r.Context())
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

// Promotion: transfer active session from anon to named user, and make named
// user the active user.
func (s *sessions) Promote(ctx context.Context, username string) bool {
	// promote anon user to auth user
	user, err := s.ActiveUserService().Promote(ctx, anon, *guser.Login)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
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

	user.ActiveSession.Data[otf.FlashSessionKey] = Flash{
		Type:    t,
		Message: fmt.Sprint(convertStringSliceToInterfaceSlice(msg)...),
	}

	if err := s.ActiveUserService.UpdateSession(r.Context(), user.ActiveSession); err != nil {
		return fmt.Errorf("saving flash message in session backend: %w", err)
	}

	return nil
}

// ActiveUser is the active user session for the current request. Provides
// methods for interacting with the active session.
type ActiveUser struct {
	*otf.User
	ActiveSession *otf.Session
}

// ActiveUserService wraps the user service, keeping track of the active
// session.
type ActiveUserService struct {
	otf.UserService
}

func (s *ActiveUserService) NewAnonymousSession(ctx context.Context) (*ActiveUser, error) {
	user, session, err := s.UserService.NewAnonymousSession(ctx)
	if err != nil {
		return nil, err
	}
	return &ActiveUser{User: user, ActiveSession: session}, nil
}

func (s *ActiveUserService) Get(ctx context.Context, token string) (*ActiveUser, error) {
	user, err := s.UserService.Get(ctx, otf.UserSpecifier{Token: &token})
	if err != nil {
		return nil, err
	}
	return &ActiveUser{User: user, ActiveSession: getActiveSession(user, token)}, nil
}

func getActiveSession(user *otf.User, token string) *otf.Session {
	for _, session := range user.Sessions {
		if session.Token == token {
			return session
		}
	}
	panic("no active session found")
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
