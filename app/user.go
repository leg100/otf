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

// Login logs a user into the system by linking their current session with their
// user account. If the user account does not exist it is created. Note:
// authentication should be handled by the caller.
func (s UserService) Login(ctx context.Context, opts otf.UserLoginOptions) error {
	user, err := s.get(ctx, opts.Username)
	if err == otf.ErrResourceNotFound {
		user, err = s.create(ctx, opts)
	} else if err != nil {
		s.Error(err, "retrieving user", "username", opts.Username)
		return err
	}

	if err := s.db.LinkSession(ctx, opts.SessionToken, user.ID); err != nil {
		s.Error(err, "user login", "username", opts.Username)
		return err
	}

	s.Info("user logged in", "username", opts.Username)

	return nil
}

func (s UserService) Get(ctx context.Context, username string) (*otf.User, error) {
	user, err := s.get(ctx, username)
	if err != nil {
		s.Error(err, "retrieving user", "username", username)
		return nil, err
	}

	s.V(2).Info("retrieved user", "username", username)

	return user, nil
}

func (s UserService) create(ctx context.Context, opts otf.UserLoginOptions) (*otf.User, error) {
	user := otf.NewUser(opts)

	if err := s.db.Create(ctx, user); err != nil {
		s.Error(err, "creating user", "username", opts.Username)
		return nil, err
	}
	s.Info("created user", "username", opts.Username)

	return user, nil
}

func (s UserService) get(ctx context.Context, username string) (*otf.User, error) {
	return s.db.Get(ctx, username)
}
