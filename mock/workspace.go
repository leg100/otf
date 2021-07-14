package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.WorkspaceService = (*WorkspaceService)(nil)

type WorkspaceService struct {
	CreateWorkspaceFn     func(org string, opts *tfe.WorkspaceCreateOptions) (*ots.Workspace, error)
	UpdateWorkspaceFn     func(name, org string, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error)
	UpdateWorkspaceByIDFn func(id string, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error)
	GetWorkspaceFn        func(name, org string) (*ots.Workspace, error)
	GetWorkspaceByIDFn    func(id string) (*ots.Workspace, error)
	ListWorkspaceFn       func(org string, opts tfe.WorkspaceListOptions) (*ots.WorkspaceList, error)
	DeleteWorkspaceFn     func(name, org string) error
	DeleteWorkspaceByIDFn func(id string) error
	LockWorkspaceFn       func(id string, opts tfe.WorkspaceLockOptions) (*ots.Workspace, error)
	UnlockWorkspaceFn     func(id string) (*ots.Workspace, error)
}

func (s WorkspaceService) Create(org string, opts *tfe.WorkspaceCreateOptions) (*ots.Workspace, error) {
	return s.CreateWorkspaceFn(org, opts)
}

func (s WorkspaceService) Update(name, org string, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error) {
	return s.UpdateWorkspaceFn(name, org, opts)
}

func (s WorkspaceService) UpdateByID(id string, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error) {
	return s.UpdateWorkspaceByIDFn(id, opts)
}

func (s WorkspaceService) Get(name, org string) (*ots.Workspace, error) {
	return s.GetWorkspaceFn(name, org)
}

func (s WorkspaceService) GetByID(id string) (*ots.Workspace, error) {
	return s.GetWorkspaceByIDFn(id)
}

func (s WorkspaceService) List(org string, opts tfe.WorkspaceListOptions) (*ots.WorkspaceList, error) {
	return s.ListWorkspaceFn(org, opts)
}

func (s WorkspaceService) Delete(name, org string) error {
	return s.DeleteWorkspaceFn(name, org)
}

func (s WorkspaceService) DeleteByID(id string) error {
	return s.DeleteWorkspaceByIDFn(id)
}

func (s WorkspaceService) Lock(id string, opts tfe.WorkspaceLockOptions) (*ots.Workspace, error) {
	return s.LockWorkspaceFn(id, opts)
}

func (s WorkspaceService) Unlock(id string) (*ots.Workspace, error) {
	return s.UnlockWorkspaceFn(id)
}

func NewWorkspace(name, id, org string) *ots.Workspace {
	return &ots.Workspace{
		ExternalID: id,
		Name:       name,
		Organization: &ots.Organization{
			Name: org,
		},
		Permissions:     &tfe.WorkspacePermissions{},
		TriggerPrefixes: []string{},
		VCSRepo:         &tfe.VCSRepo{},
	}
}

func NewWorkspaceList(name, id, org string, opts tfe.WorkspaceListOptions) *ots.WorkspaceList {
	return &ots.WorkspaceList{
		Items: []*ots.Workspace{
			NewWorkspace(name, id, org),
		},
		Pagination: ots.NewPagination(opts.ListOptions, 1),
	}
}
