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

// SwapUser transfers the active session to a user with the given username and
// attaches that user on the context, replacing the existing user. Returns the
// new context with the new user attached.
func (s *sessions) SwapUser(ctx context.Context, user *otf.User) (context.Context, error) {
	existing := s.getUserFromContext(ctx)

	_, err := s.ActiveUserService.TransferSession(ctx, existing.ActiveSession.Token, existing.User, user)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, userCtxKey, user), nil
}

// Destroy deletes the current session, removing the current user.
func (s *sessions) Destroy(ctx context.Context) error {
	user := s.getUserFromContext(ctx)
	return s.ActiveUserService.DeleteSession(ctx, user.ActiveSession.Token)
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
func (s *sessions) PopFlash(r *http.Request) *Flash {
	// TODO: need to *remove* session data entry.
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

	flash = Flash{
		Type:    t,
		Message: fmt.Sprint(convertStringSliceToInterfaceSlice(msg)...),
	}

	if err := s.UserService.UpdateSessionData(r.Context(), user.ActiveSession.Token, otf.FlashSessionKey, flash); err != nil {
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

// NewAnonymousSession returns the anonymous user with a new session.
func (s *ActiveUserService) NewAnonymousSession(r *http.Request) (*ActiveUser, error) {
	anon, err := s.UserService.GetAnonymous(r.Context())
	if err != nil {
		return nil, err
	}

	session, err := s.UserService.CreateSession(r.Context(), anon, newSessionData(r))
	if err != nil {
		return nil, err
	}

	return &ActiveUser{User: anon, ActiveSession: session}, nil
}

func (s *ActiveUserService) Get(ctx context.Context, token string) (*ActiveUser, error) {
	user, err := s.UserService.Get(ctx, otf.UserSpecifier{Token: &token})
	if err != nil {
		return nil, err
	}
	return &ActiveUser{User: user, ActiveSession: getActiveSession(user, token)}, nil
}

func newSessionData(r *http.Request) (otf.SessionData, error) {
	addr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, err
	}

	data := otf.SessionData{
		otf.AddressSessionKey: addr,
	}

	return data, nil
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
