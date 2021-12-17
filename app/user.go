package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.UserService = (*UserService)(nil)

type UserService struct {
	db otf.SessionStore

	logr.Logger
}

func NewUserService(db otf.UserStore, logger logr.Logger, os otf.OrganizationService, es otf.EventService) *UserService {
	return &UserService{
		db:     db,
		es:     es,
		os:     os,
		Logger: logger,
	}
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

func (s UserService) Delete(ctx context.Context, spec otf.UserSpecifier) error {
	// Get workspace so we can publish it in an event after we delete it
	ws, err := s.db.Get(spec)
	if err != nil {
		return err
	}

	if err := s.db.Delete(spec); err != nil {
		s.Error(err, "deleting workspace", "id", ws.ID, "name", ws.Name)
		return err
	}

	s.es.Publish(otf.Event{Type: otf.EventUserDeleted, Payload: ws})

	s.V(0).Info("deleted workspace", "id", ws.ID, "name", ws.Name)

	return nil
}

func (s UserService) Lock(ctx context.Context, id string, _ otf.UserLockOptions) (*otf.User, error) {
	spec := otf.UserSpecifier{ID: &id}

	return s.db.Update(spec, func(ws *otf.User) (err error) {
		return ws.ToggleLock(true)
	})
}

func (s UserService) Unlock(ctx context.Context, id string) (*otf.User, error) {
	spec := otf.UserSpecifier{ID: &id}

	return s.db.Update(spec, func(ws *otf.User) (err error) {
		return ws.ToggleLock(false)
	})
}
