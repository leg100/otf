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

func (s WorkspaceService) Create(ctx context.Context, opts otf.WorkspaceCreateOptions) (*otf.Workspace, error) {
	if err := opts.Valid(); err != nil {
		return nil, err
	}

	org, err := s.os.Get(ctx, opts.OrganizationName)
	if err != nil {
		return nil, err
	}

	ws := otf.NewWorkspace(opts, org)

	_, err = s.db.Create(ws)
	if err != nil {
		s.Error(err, "creating workspace", "id", ws.ID(), "name", ws.Name())
		return nil, err
	}

	s.V(0).Info("created workspace", "id", ws.ID(), "name", ws.Name())

	s.es.Publish(otf.Event{Type: otf.EventWorkspaceCreated, Payload: ws})

	return ws, nil
}

func (s WorkspaceService) Update(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	if err := opts.Valid(); err != nil {
		s.Error(err, "updating workspace")
		return nil, err
	}

	ws, err := s.db.Update(spec, func(ws *otf.Workspace) error {
		return ws.UpdateWithOptions(ctx, opts)
	})
	if err != nil {
		s.Error(err, "updating workspace", spec.LogInfo()...)
		return nil, err
	}

	s.V(0).Info("updated workspace", spec.LogInfo()...)

	return ws, nil
}

func (s WorkspaceService) List(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return s.db.List(opts)
}

func (s WorkspaceService) Get(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	if err := spec.Valid(); err != nil {
		s.Error(err, "retrieving workspace")
		return nil, err
	}

	ws, err := s.db.Get(spec)
	if err != nil {
		s.Error(err, "retrieving workspace", spec.LogInfo()...)
		return nil, err
	}

	s.V(2).Info("retrieved workspace", spec.LogInfo()...)

	return ws, nil
}

func (s WorkspaceService) Delete(ctx context.Context, spec otf.WorkspaceSpec) error {
	// Get workspace so we can publish it in an event after we delete it
	ws, err := s.db.Get(spec)
	if err != nil {
		return err
	}

	if err := s.db.Delete(spec); err != nil {
		s.Error(err, "deleting workspace", "id", ws.ID(), "name", ws.Name())
		return err
	}

	s.es.Publish(otf.Event{Type: otf.EventWorkspaceDeleted, Payload: ws})

	s.V(0).Info("deleted workspace", "id", ws.ID(), "name", ws.Name())

	return nil
}

func (s WorkspaceService) Lock(ctx context.Context, spec otf.WorkspaceSpec, _ otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	ws, err := s.db.Update(spec, func(ws *otf.Workspace) error {
		return ws.ToggleLock(true)
	})
	if err != nil {
		s.Error(err, "locking workspace", spec.LogInfo()...)
		return nil, err
	}

	s.V(1).Info("locked workspace", spec.LogInfo()...)

	return ws, nil
}

func (s WorkspaceService) Unlock(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	ws, err := s.db.Update(spec, func(ws *otf.Workspace) error {
		return ws.ToggleLock(false)
	})
	if err != nil {
		s.Error(err, "unlocking workspace", spec.LogInfo()...)
		return nil, err
	}

	s.V(1).Info("unlocked workspace", spec.LogInfo()...)

	return ws, nil
}
