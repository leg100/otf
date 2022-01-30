package html

import (
	"context"
	"net/http"

	"github.com/leg100/otf"
)

// ActiveUser is the active user session for the current request. Provides
// methods for interacting with the active session.
type ActiveUser struct {
	*otf.User
	*otf.Session
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

	data, err := newSessionData(r)
	if err != nil {
		return nil, err
	}

	session, err := s.UserService.CreateSession(r.Context(), anon, data)
	if err != nil {
		return nil, err
	}

	return &ActiveUser{User: anon, Session: session}, nil
}

func (s *ActiveUserService) Get(ctx context.Context, token string) (*ActiveUser, error) {
	user, err := s.UserService.Get(ctx, otf.UserSpec{SessionToken: &token})
	if err != nil {
		return nil, err
	}
	return &ActiveUser{User: user, Session: getActiveSession(user, token)}, nil
}

func getActiveSession(user *otf.User, token string) *otf.Session {
	for _, session := range user.Sessions {
		if session.Token == token {
			return session
		}
	}
	panic("no active session found")
}
