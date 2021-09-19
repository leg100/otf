package app

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
)

var _ otf.WorkspaceService = (*WorkspaceService)(nil)

type WorkspaceService struct {
	db otf.WorkspaceStore
	os otf.OrganizationService
	es otf.EventService
}

func NewWorkspaceService(db otf.WorkspaceStore, os otf.OrganizationService, es otf.EventService) *WorkspaceService {
	return &WorkspaceService{
		db: db,
		es: es,
		os: os,
	}
}

func (s WorkspaceService) Create(orgName string, opts *tfe.WorkspaceCreateOptions) (*otf.Workspace, error) {
	org, err := s.os.Get(orgName)
	if err != nil {
		return nil, err
	}

	ws := otf.NewWorkspace(opts, org)

	ws, err = s.db.Create(ws)
	if err != nil {
		return nil, err
	}

	s.es.Publish(otf.Event{Type: otf.WorkspaceCreated, Payload: ws})

	return ws, nil
}

func (s WorkspaceService) Update(spec otf.WorkspaceSpecifier, opts *tfe.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	return s.db.Update(spec, func(ws *otf.Workspace) (err error) {
		_, err = otf.UpdateWorkspace(ws, opts)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s WorkspaceService) List(opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return s.db.List(opts)
}

func (s WorkspaceService) Get(spec otf.WorkspaceSpecifier) (*otf.Workspace, error) {
	return s.db.Get(spec)
}

func (s WorkspaceService) Delete(spec otf.WorkspaceSpecifier) error {
	// Get workspace so we can publish it in an event after we delete it
	ws, err := s.db.Get(spec)
	if err != nil {
		return err
	}

	if err := s.db.Delete(spec); err != nil {
		return err
	}

	s.es.Publish(otf.Event{Type: otf.WorkspaceDeleted, Payload: ws})

	return nil
}

func (s WorkspaceService) Lock(id string, _ tfe.WorkspaceLockOptions) (*otf.Workspace, error) {
	spec := otf.WorkspaceSpecifier{ID: &id}

	return s.db.Update(spec, func(ws *otf.Workspace) (err error) {
		return ws.ToggleLock(true)
	})
}

func (s WorkspaceService) Unlock(id string) (*otf.Workspace, error) {
	spec := otf.WorkspaceSpecifier{ID: &id}

	return s.db.Update(spec, func(ws *otf.Workspace) (err error) {
		return ws.ToggleLock(false)
	})
}
