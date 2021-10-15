package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.WorkspaceService = (*WorkspaceService)(nil)

type WorkspaceService struct {
	db otf.WorkspaceStore
	os otf.OrganizationService
	es otf.EventService

	logr.Logger
}

func NewWorkspaceService(db otf.WorkspaceStore, logger logr.Logger, os otf.OrganizationService, es otf.EventService) *WorkspaceService {
	return &WorkspaceService{
		db:     db,
		es:     es,
		os:     os,
		Logger: logger,
	}
}

func (s WorkspaceService) Create(ctx context.Context, orgName string, opts otf.WorkspaceCreateOptions) (*otf.Workspace, error) {
	if err := opts.Valid(); err != nil {
		return nil, err
	}

	org, err := s.os.Get(orgName)
	if err != nil {
		return nil, err
	}

	ws := otf.NewWorkspace(opts, org)

	_, err = s.db.Create(ws)
	if err != nil {
		s.Error(err, "creating workspace", "id", ws.ID)
		return nil, err
	}

	s.Info("created workspace", "id", ws.ID)

	s.es.Publish(otf.Event{Type: otf.WorkspaceCreated, Payload: ws})

	return ws, nil
}

func (s WorkspaceService) Update(ctx context.Context, spec otf.WorkspaceSpecifier, opts otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	if err := opts.Valid(); err != nil {
		return nil, err
	}

	return s.db.Update(spec, func(ws *otf.Workspace) (err error) {
		_, err = otf.UpdateWorkspace(ws, opts)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s WorkspaceService) List(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return s.db.List(opts)
}

func (s WorkspaceService) Get(ctx context.Context, spec otf.WorkspaceSpecifier) (*otf.Workspace, error) {
	return s.db.Get(spec)
}

func (s WorkspaceService) Delete(ctx context.Context, spec otf.WorkspaceSpecifier) error {
	// Get workspace so we can publish it in an event after we delete it
	ws, err := s.db.Get(spec)
	if err != nil {
		return err
	}

	if err := s.db.Delete(spec); err != nil {
		s.Error(err, "deleting workspace")
		return err
	}

	s.es.Publish(otf.Event{Type: otf.WorkspaceDeleted, Payload: ws})

	return nil
}

func (s WorkspaceService) Lock(ctx context.Context, id string, _ otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	spec := otf.WorkspaceSpecifier{ID: &id}

	return s.db.Update(spec, func(ws *otf.Workspace) (err error) {
		return ws.ToggleLock(true)
	})
}

func (s WorkspaceService) Unlock(ctx context.Context, id string) (*otf.Workspace, error) {
	spec := otf.WorkspaceSpecifier{ID: &id}

	return s.db.Update(spec, func(ws *otf.Workspace) (err error) {
		return ws.ToggleLock(false)
	})
}
