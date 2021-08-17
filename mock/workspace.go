package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.WorkspaceService = (*WorkspaceService)(nil)

type WorkspaceService struct {
	CreateWorkspaceFn func(org string, opts *tfe.WorkspaceCreateOptions) (*ots.Workspace, error)
	UpdateWorkspaceFn func(spec ots.WorkspaceSpecifier, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error)
	GetWorkspaceFn    func(spec ots.WorkspaceSpecifier) (*ots.Workspace, error)
	ListWorkspaceFn   func(opts ots.WorkspaceListOptions) (*ots.WorkspaceList, error)
	DeleteWorkspaceFn func(spec ots.WorkspaceSpecifier) error
	LockWorkspaceFn   func(id string, opts tfe.WorkspaceLockOptions) (*ots.Workspace, error)
	UnlockWorkspaceFn func(id string) (*ots.Workspace, error)
}

func (s WorkspaceService) Create(org string, opts *tfe.WorkspaceCreateOptions) (*ots.Workspace, error) {
	return s.CreateWorkspaceFn(org, opts)
}

func (s WorkspaceService) Update(spec ots.WorkspaceSpecifier, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error) {
	return s.UpdateWorkspaceFn(spec, opts)
}

func (s WorkspaceService) Get(spec ots.WorkspaceSpecifier) (*ots.Workspace, error) {
	return s.GetWorkspaceFn(spec)
}

func (s WorkspaceService) List(opts ots.WorkspaceListOptions) (*ots.WorkspaceList, error) {
	return s.ListWorkspaceFn(opts)
}

func (s WorkspaceService) Delete(spec ots.WorkspaceSpecifier) error {
	return s.DeleteWorkspaceFn(spec)
}

func (s WorkspaceService) Lock(id string, opts tfe.WorkspaceLockOptions) (*ots.Workspace, error) {
	return s.LockWorkspaceFn(id, opts)
}

func (s WorkspaceService) Unlock(id string) (*ots.Workspace, error) {
	return s.UnlockWorkspaceFn(id)
}

func NewWorkspace(name, id, org string) *ots.Workspace {
	return &ots.Workspace{
		ID:   id,
		Name: name,
		Organization: &ots.Organization{
			Name: org,
		},
		Permissions:     &tfe.WorkspacePermissions{},
		TriggerPrefixes: []string{},
		VCSRepo:         &tfe.VCSRepo{},
	}
}

func NewWorkspaceList(name, id, org string, opts ots.WorkspaceListOptions) *ots.WorkspaceList {
	return &ots.WorkspaceList{
		Items: []*ots.Workspace{
			NewWorkspace(name, id, org),
		},
		Pagination: ots.NewPagination(opts.ListOptions, 1),
	}
}
