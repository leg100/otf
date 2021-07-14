package app

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.WorkspaceService = (*WorkspaceService)(nil)

type WorkspaceService struct {
	db ots.WorkspaceRepository
	os ots.OrganizationService
}

func NewWorkspaceService(db ots.WorkspaceRepository, os ots.OrganizationService) *WorkspaceService {
	return &WorkspaceService{
		db: db,
		os: os,
	}
}

func (s WorkspaceService) Create(orgName string, opts *tfe.WorkspaceCreateOptions) (*ots.Workspace, error) {
	org, err := s.os.Get(orgName)
	if err != nil {
		return nil, err
	}

	ws := ots.NewWorkspace(opts, org)

	_, err = s.db.Create(ws)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (s WorkspaceService) Update(name, orgName string, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error) {
	spec := ots.WorkspaceSpecifier{Name: &name, OrganizationName: &orgName}

	return s.db.Update(spec, func(ws *ots.Workspace) (err error) {
		_, err = ots.UpdateWorkspace(ws, opts)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s WorkspaceService) UpdateByID(id string, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error) {
	spec := ots.WorkspaceSpecifier{ID: &id}

	return s.db.Update(spec, func(ws *ots.Workspace) (err error) {
		_, err = ots.UpdateWorkspace(ws, opts)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s WorkspaceService) List(orgName string, opts tfe.WorkspaceListOptions) (*ots.WorkspaceList, error) {
	return s.db.List(orgName, ots.WorkspaceListOptions{ListOptions: opts.ListOptions, Prefix: opts.Search})
}

func (s WorkspaceService) Get(name, orgName string) (*ots.Workspace, error) {
	return s.db.Get(ots.WorkspaceSpecifier{Name: &name, OrganizationName: &orgName})
}

func (s WorkspaceService) GetByID(id string) (*ots.Workspace, error) {
	return s.db.Get(ots.WorkspaceSpecifier{ID: &id})
}

func (s WorkspaceService) Delete(name, orgName string) error {
	return s.db.Delete(ots.WorkspaceSpecifier{Name: &name, OrganizationName: &orgName})
}

func (s WorkspaceService) DeleteByID(id string) error {
	return s.db.Delete(ots.WorkspaceSpecifier{ID: &id})
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
