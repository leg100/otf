package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
)

var _ otf.WorkspaceService = (*WorkspaceService)(nil)

type WorkspaceService struct {
	CreateWorkspaceFn func(org string, opts *tfe.WorkspaceCreateOptions) (*otf.Workspace, error)
	UpdateWorkspaceFn func(spec otf.WorkspaceSpecifier, opts *tfe.WorkspaceUpdateOptions) (*otf.Workspace, error)
	GetWorkspaceFn    func(spec otf.WorkspaceSpecifier) (*otf.Workspace, error)
	ListWorkspaceFn   func(opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error)
	DeleteWorkspaceFn func(spec otf.WorkspaceSpecifier) error
	LockWorkspaceFn   func(id string, opts tfe.WorkspaceLockOptions) (*otf.Workspace, error)
	UnlockWorkspaceFn func(id string) (*otf.Workspace, error)
}

func (s WorkspaceService) Create(org string, opts *tfe.WorkspaceCreateOptions) (*otf.Workspace, error) {
	return s.CreateWorkspaceFn(org, opts)
}

func (s WorkspaceService) Update(spec otf.WorkspaceSpecifier, opts *tfe.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	return s.UpdateWorkspaceFn(spec, opts)
}

func (s WorkspaceService) Get(spec otf.WorkspaceSpecifier) (*otf.Workspace, error) {
	return s.GetWorkspaceFn(spec)
}

func (s WorkspaceService) List(opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return s.ListWorkspaceFn(opts)
}

func (s WorkspaceService) Delete(spec otf.WorkspaceSpecifier) error {
	return s.DeleteWorkspaceFn(spec)
}

func (s WorkspaceService) Lock(id string, opts tfe.WorkspaceLockOptions) (*otf.Workspace, error) {
	return s.LockWorkspaceFn(id, opts)
}

func (s WorkspaceService) Unlock(id string) (*otf.Workspace, error) {
	return s.UnlockWorkspaceFn(id)
}

func NewWorkspace(name, id, org string) *otf.Workspace {
	return &otf.Workspace{
		ID:   id,
		Name: name,
		Organization: &otf.Organization{
			Name: org,
		},
		Permissions:     &tfe.WorkspacePermissions{},
		TriggerPrefixes: []string{},
		VCSRepo:         &tfe.VCSRepo{},
	}
}

func NewWorkspaceList(name, id, org string, opts otf.WorkspaceListOptions) *otf.WorkspaceList {
	return &otf.WorkspaceList{
		Items: []*otf.Workspace{
			NewWorkspace(name, id, org),
		},
		Pagination: otf.NewPagination(opts.ListOptions, 1),
	}
}
