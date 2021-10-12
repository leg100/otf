package mock

import (
	"context"

	"github.com/leg100/otf"
)

var _ otf.WorkspaceService = (*WorkspaceService)(nil)

type WorkspaceService struct {
	CreateWorkspaceFn func(org string, opts otf.WorkspaceCreateOptions) (*otf.Workspace, error)
	UpdateWorkspaceFn func(spec otf.WorkspaceSpecifier, opts otf.WorkspaceUpdateOptions) (*otf.Workspace, error)
	GetWorkspaceFn    func(spec otf.WorkspaceSpecifier) (*otf.Workspace, error)
	ListWorkspaceFn   func(opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error)
	DeleteWorkspaceFn func(spec otf.WorkspaceSpecifier) error
	LockWorkspaceFn   func(id string, opts otf.WorkspaceLockOptions) (*otf.Workspace, error)
	UnlockWorkspaceFn func(id string) (*otf.Workspace, error)
}

func (s WorkspaceService) Create(ctx context.Context, org string, opts otf.WorkspaceCreateOptions) (*otf.Workspace, error) {
	return s.CreateWorkspaceFn(org, opts)
}

func (s WorkspaceService) Update(ctx context.Context, spec otf.WorkspaceSpecifier, opts otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	return s.UpdateWorkspaceFn(spec, opts)
}

func (s WorkspaceService) Get(ctx context.Context, spec otf.WorkspaceSpecifier) (*otf.Workspace, error) {
	return s.GetWorkspaceFn(spec)
}

func (s WorkspaceService) List(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return s.ListWorkspaceFn(opts)
}

func (s WorkspaceService) Delete(ctx context.Context, spec otf.WorkspaceSpecifier) error {
	return s.DeleteWorkspaceFn(spec)
}

func (s WorkspaceService) Lock(ctx context.Context, id string, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	return s.LockWorkspaceFn(id, opts)
}

func (s WorkspaceService) Unlock(ctx context.Context, id string) (*otf.Workspace, error) {
	return s.UnlockWorkspaceFn(id)
}

func NewWorkspace(name, id, org string) *otf.Workspace {
	return &otf.Workspace{
		ID:   id,
		Name: name,
		Organization: &otf.Organization{
			Name: org,
		},
		TriggerPrefixes: []string{},
		VCSRepo:         &otf.VCSRepo{},
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
