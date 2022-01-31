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
		s.Error(err, "creating user", "username", user.Username)
		return nil, err
	}

	s.V(1).Info("created user", "username", user.Username)

	return user, nil
}

func (s UserService) EnsureCreated(ctx context.Context, username string) (*otf.User, error) {
	user, err := s.db.Get(ctx, otf.UserSpec{Username: &username})
	if err == nil {
		return user, nil
	}

	if err != otf.ErrResourceNotFound {
		s.Error(err, "retrieving user", "username", username)
		return nil, err
	}

	return s.Create(ctx, username)
}

func (s UserService) Update(ctx context.Context, username string, updated *otf.User) error {
	if err := s.db.Update(ctx, otf.UserSpec{Username: &username}, updated); err != nil {
		s.Error(err, "updating user", "username", username)
		return err
	}

	s.V(1).Info("updated user", "username", username)

	return nil
}

// CreateSession creates a session and adds it to the user.
func (s UserService) CreateSession(ctx context.Context, user *otf.User, data *otf.SessionData) (*otf.Session, error) {
	session, err := user.AttachNewSession(data)
	if err != nil {
		s.Error(err, "attaching session", "username", user.Username)
		return nil, err
	}

	if err := s.db.CreateSession(ctx, session); err != nil {
		s.Error(err, "creating session", "username", user.Username)
		return nil, err
	}

	s.V(1).Info("created session", "username", user.Username)

	return session, nil
}

func (s UserService) Get(ctx context.Context, spec otf.UserSpec) (*otf.User, error) {
	user, err := s.db.Get(ctx, spec)
	if err != nil {
		s.Error(err, "retrieving user", spec.KeyValue()...)
		return nil, err
	}

	s.V(2).Info("retrieved user", "username", user.Username)

	return user, nil
}

func (s UserService) GetAnonymous(ctx context.Context) (*otf.User, error) {
	return s.Get(ctx, otf.UserSpec{Username: otf.String(otf.AnonymousUsername)})
}

// UpdateSession updates a user session.
func (s UserService) UpdateSession(ctx context.Context, user *otf.User, session *otf.Session) error {
	err := s.db.UpdateSession(ctx, session.Token, session)
	if err != nil {
		s.Error(err, "updating session", "username", user.Username)
		return err
	}

	s.V(1).Info("updated session", "username", user.Username)

	return nil
}

func (s UserService) DeleteSession(ctx context.Context, token string) error {
	// Retrieve user purely for logging purposes
	user, err := s.Get(ctx, otf.UserSpec{SessionToken: &token})
	if err != nil {
		return err
	}

	if err := s.db.DeleteSession(ctx, token); err != nil {
		s.Error(err, "deleting session", "username", user.Username)
		return err
	}

	s.V(1).Info("deleted session", "username", user.Username)

	return nil
}

// CreateToken creates a user token.
func (s UserService) CreateToken(ctx context.Context, user *otf.User, opts *otf.TokenCreateOptions) (*otf.Token, error) {
	token, err := otf.NewToken(user.ID, opts.Description)
	if err != nil {
		s.Error(err, "constructing token", "username", user.Username)
		return nil, err
	}

	if err := s.db.CreateToken(ctx, token); err != nil {
		s.Error(err, "creating token", "username", user.Username)
		return nil, err
	}

	s.V(1).Info("created token", "username", user.Username)

	return token, nil
}

func (s UserService) DeleteToken(ctx context.Context, id string) error {
	// Retrieve user purely for logging purposes
	user, err := s.Get(ctx, otf.UserSpec{AuthenticationTokenID: &id})
	if err != nil {
		return err
	}

	if err := s.db.DeleteToken(ctx, id); err != nil {
		s.Error(err, "deleting token", "username", user.Username)
		return err
	}

	s.V(1).Info("deleted token", "username", user.Username)

	return nil
}
