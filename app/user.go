package app

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.UserService = (*UserService)(nil)

type UserService struct {
	db otf.UserStore

	logr.Logger
}

func NewUserService(logger logr.Logger, db otf.DB) *UserService {
	return &UserService{
		db:     db.UserStore(),
		Logger: logger,
	}
}

// NewAnonymousSession creates a new session for the anonymous user (all
// sessions start their life as anonymous sessions) and returns the anonymous
// user and the new session.
func (s UserService) NewAnonymousSession(ctx context.Context) (*otf.User, *otf.Session, error) {
	anon, err := s.db.Get(ctx, otf.UserSpecifier{Username: otf.String(otf.AnonymousUsername)})
	if err != nil {
		s.Error(err, "retrieving user", "username", anon.Username)
		return nil, nil, err
	}

	session, err := anon.AttachNewSession()
	if err != nil {
		s.Error(err, "attaching session", "username", anon.Username)
		return nil, nil, err
	}

	if err := s.db.CreateSession(ctx, session); err != nil {
		s.Error(err, "creating session", "username", anon.Username)
		return nil, nil, err
	}

	return anon, session, nil
}

// Promote transfers the anonymous user's session to a named user with the given
// username. If no such user exists, a new user is created.
func (s UserService) Promote(ctx context.Context, anon *otf.User, username string) (*otf.User, error) {
	if total := len(anon.Sessions); total != 0 {
		return nil, fmt.Errorf("an anonymous user must always have one session but %d were found", total)
	}

	session := anon.Sessions[0]

	// Get named user; if not exist create one
	named, err := s.db.Get(ctx, otf.UserSpecifier{Username: &username})
	if err == otf.ErrResourceNotFound {
		named, err = s.create(ctx, username)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		s.Error(err, "retrieving user", "username", username)
		return nil, err
	}

	// Transfer session from anon to named user
	if err := s.db.LinkSession(ctx, session, named); err != nil {
		s.Error(err, "linking session", "username", username)
		return nil, err
	}

	s.Info("promoted user", "username", username)

	return named, nil
}

func (s UserService) Get(ctx context.Context, spec otf.UserSpecifier) (*otf.User, error) {
	user, err := s.db.Get(ctx, spec)
	if err != nil {
		s.Error(err, "retrieving user", spec.KeyValue()...)
		return nil, err
	}

	s.V(2).Info("retrieved user", "username", user.Username)

	return user, nil
}

// TransferSession transfers a session from one user to another.
func (s UserService) TransferSession(ctx context.Context, session *otf.Session, from, to *otf.User) error {
	session.UserID = to.ID

	if err := s.db.UpdateSession(ctx, session); err != nil {
		s.Error(err, "transferring session", "from", from, "to", to)
		return err
	}

	s.V(1).Info("transferred session", "from", from, "to", to)

	return nil
}

func (s UserService) UpdateSession(ctx context.Context, user *otf.User, session *otf.Session) error {
	if err := s.db.UpdateSession(ctx, session); err != nil {
		s.Error(err, "updating session", "username", user.Username)
		return err
	}

	s.V(1).Info("updated session", "username", user.Username)

	return nil
}

func (s UserService) RevokeSession(ctx context.Context, token, username string) error {
	if err := s.db.RevokeSession(ctx, token, username); err != nil {
		s.Error(err, "revoking session", "username", username)
		return err
	}

	s.V(1).Info("revoked session", "username", username)

	return nil
}

func (s UserService) create(ctx context.Context, username string) (*otf.User, error) {
	user := otf.NewUser(username)

	if err := s.db.Create(ctx, user); err != nil {
		s.Error(err, "creating user", "username", username)
		return nil, err
	}
	s.Info("created user", "username", user.Username)

	return user, nil
}
