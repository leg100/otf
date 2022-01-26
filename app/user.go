package app

import (
	"context"

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

func (s UserService) Create(ctx context.Context, username string) (*otf.User, error) {
	user := otf.NewUser(username)

	if err := s.db.Create(ctx, user); err != nil {
		s.Error(err, "creating user", "user", user)
		return nil, err
	}

	s.V(1).Info("created user", "user", user)

	return user, nil
}

// CreateSession creates a session and adds it to the user.
func (s UserService) CreateSession(ctx context.Context, user *otf.User, data *otf.SessionData) (*otf.Session, error) {
	session, err := user.AttachNewSession(data)
	if err != nil {
		s.Error(err, "attaching session", "user", user)
		return nil, err
	}

	if err := s.db.CreateSession(ctx, session); err != nil {
		s.Error(err, "creating session", "user", user)
		return nil, err
	}

	s.V(1).Info("created session", "user", user)

	return session, nil
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

func (s UserService) GetAnonymous(ctx context.Context) (*otf.User, error) {
	return s.Get(ctx, otf.UserSpecifier{Username: otf.String(otf.AnonymousUsername)})
}

// UpdateSession updates a user session.
func (s UserService) UpdateSession(ctx context.Context, user *otf.User, session *otf.Session) error {
	err := s.db.UpdateSession(ctx, session.Token, session)
	if err != nil {
		s.Error(err, "updating session", "user", user)
		return err
	}

	s.V(1).Info("updated session", "user", user)

	return nil
}

func (s UserService) DeleteSession(ctx context.Context, token string) error {
	// Retrieve user purely for logging purposes
	user, err := s.Get(ctx, otf.UserSpecifier{Token: &token})
	if err != nil {
		return err
	}

	if err := s.db.DeleteSession(ctx, token); err != nil {
		s.Error(err, "revoking session", "username", user)
		return err
	}

	s.V(1).Info("revoked session", "username", user)

	return nil
}
