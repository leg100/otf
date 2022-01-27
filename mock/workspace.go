package mock

import (
	"context"

	"github.com/leg100/otf"
)

var _ otf.WorkspaceService = (*WorkspaceService)(nil)

type WorkspaceService struct {
	CreateWorkspaceFn func(opts otf.WorkspaceCreateOptions) (*otf.Workspace, error)
	UpdateWorkspaceFn func(spec otf.WorkspaceSpec, opts otf.WorkspaceUpdateOptions) (*otf.Workspace, error)
	GetWorkspaceFn    func(spec otf.WorkspaceSpec) (*otf.Workspace, error)
	ListWorkspaceFn   func(opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error)
	DeleteWorkspaceFn func(spec otf.WorkspaceSpec) error
	LockWorkspaceFn   func(spec otf.WorkspaceSpec, opts otf.WorkspaceLockOptions) (*otf.Workspace, error)
	UnlockWorkspaceFn func(spec otf.WorkspaceSpec) (*otf.Workspace, error)
}

func (s WorkspaceService) Create(ctx context.Context, opts otf.WorkspaceCreateOptions) (*otf.Workspace, error) {
	return s.CreateWorkspaceFn(opts)
}

func (s WorkspaceService) Update(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	return s.UpdateWorkspaceFn(spec, opts)
}

func (s WorkspaceService) Get(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	return s.GetWorkspaceFn(spec)
}

func (s WorkspaceService) List(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return s.ListWorkspaceFn(opts)
}

func (s WorkspaceService) Delete(ctx context.Context, spec otf.WorkspaceSpec) error {
	return s.DeleteWorkspaceFn(spec)
}

func (s WorkspaceService) Lock(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	return s.LockWorkspaceFn(spec, opts)
}

func (s WorkspaceService) Unlock(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	return s.UnlockWorkspaceFn(spec)
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
