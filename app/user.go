package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.UserService = (*UserService)(nil)

type UserService struct {
	db       otf.UserStore
	sessions otf.SessionStore

	logr.Logger
}

func NewUserService(logger logr.Logger, db otf.DB) *UserService {
	return &UserService{
		db:       db.UserStore(),
		sessions: db.SessionStore(),
		Logger:   logger,
	}
}

// Login logs a user into the system. A user is created if they don't already
// exist. Note: authentication is handled upstream in the http package.
func (s UserService) Login(ctx context.Context, opts otf.UserLoginOptions) error {
	user, err := s.get(ctx, opts.Username)
	if err == otf.ErrResourceNotFound {
		user, err = s.create(ctx, opts.Username)
	} else if err != nil {
		s.Error(err, "retrieving user", "username", opts.Username)
		return err
	}

	// Associate user with session token
	_, err = s.sessions.Update(ctx, opts.SessionToken, func(session *otf.Session) error {
		session.User = user
		return nil
	})
	if err != nil {
		s.Error(err, "user login", "username", opts.Username)
		return err
	}

	s.Info("user logged in", "username", opts.Username)

	return nil
}

func (s UserService) Sessions(ctx context.Context) ([]*otf.Session, error) {
	return s.db.List(opts)
}

func (s UserService) Get(ctx context.Context, spec otf.UserSpecifier) (*otf.User, error) {
	if err := spec.Valid(); err != nil {
		s.Error(err, "retrieving workspace: invalid specifier")
		return nil, err
	}

	ws, err := s.db.Get(spec)
	if err != nil {
		s.Error(err, "retrieving workspace", "id", spec.String())
		return nil, err
	}

	s.V(2).Info("retrieved workspace", "id", spec.String())

	return ws, nil
}

func (s UserService) create(ctx context.Context, username string) (*otf.User, error) {
	user, err := s.db.Create(ctx, username)
	if err != nil {
		s.Error(err, "creating user", "username", username)
		return err
	}
	s.Info(err, "created user", "username", username)

	return user, err
}

func (s UserService) get(ctx context.Context, username string) (*otf.User, error) {
	return s.db.Get(ctx, username)
}
