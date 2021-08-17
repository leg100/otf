package app

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.WorkspaceService = (*WorkspaceService)(nil)

type WorkspaceService struct {
	db ots.WorkspaceStore
	os ots.OrganizationService
	es ots.EventService
}

func NewWorkspaceService(db ots.WorkspaceStore, os ots.OrganizationService, es ots.EventService) *WorkspaceService {
	return &WorkspaceService{
		db: db,
		es: es,
		os: os,
	}
}

func (s WorkspaceService) Create(orgName string, opts *tfe.WorkspaceCreateOptions) (*ots.Workspace, error) {
	org, err := s.os.Get(orgName)
	if err != nil {
		return nil, err
	}

	ws := ots.NewWorkspace(opts, org)

	ws, err = s.db.Create(ws)
	if err != nil {
		return nil, err
	}

	s.es.Publish(ots.Event{Type: ots.WorkspaceCreated, Payload: ws})

	return ws, nil
}

func (s WorkspaceService) Update(spec ots.WorkspaceSpecifier, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error) {
	return s.db.Update(spec, func(ws *ots.Workspace) (err error) {
		_, err = ots.UpdateWorkspace(ws, opts)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s WorkspaceService) List(opts ots.WorkspaceListOptions) (*ots.WorkspaceList, error) {
	return s.db.List(opts)
}

func (s WorkspaceService) Get(spec ots.WorkspaceSpecifier) (*ots.Workspace, error) {
	return s.db.Get(spec)
}

func (s WorkspaceService) Delete(spec ots.WorkspaceSpecifier) error {
	// Get workspace so we can publish it in an event after we delete it
	ws, err := s.db.Get(spec)
	if err != nil {
		return err
	}

	if err := s.db.Delete(spec); err != nil {
		return err
	}

	s.es.Publish(ots.Event{Type: ots.WorkspaceDeleted, Payload: ws})

	return nil
}

func (s WorkspaceService) Lock(id string, _ tfe.WorkspaceLockOptions) (*ots.Workspace, error) {
	spec := ots.WorkspaceSpecifier{ID: &id}

	return s.db.Update(spec, func(ws *ots.Workspace) (err error) {
		return ws.ToggleLock(true)
	})
}

func (s WorkspaceService) Unlock(id string) (*ots.Workspace, error) {
	spec := ots.WorkspaceSpecifier{ID: &id}

	return s.db.Update(spec, func(ws *ots.Workspace) (err error) {
		return ws.ToggleLock(false)
	})
}
