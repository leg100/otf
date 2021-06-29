package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.WorkspaceService = (*WorkspaceService)(nil)

type WorkspaceService struct {
	CreateWorkspaceFn     func(org string, opts *tfe.WorkspaceCreateOptions) (*tfe.Workspace, error)
	UpdateWorkspaceFn     func(name, org string, opts *tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error)
	UpdateWorkspaceByIDFn func(id string, opts *tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error)
	GetWorkspaceFn        func(name, org string) (*tfe.Workspace, error)
	GetWorkspaceByIDFn    func(id string) (*tfe.Workspace, error)
	ListWorkspaceFn       func(org string, opts ots.WorkspaceListOptions) (*ots.WorkspaceList, error)
	DeleteWorkspaceFn     func(name, org string) error
	DeleteWorkspaceByIDFn func(id string) error
	LockWorkspaceFn       func(id string, opts ots.WorkspaceLockOptions) (*tfe.Workspace, error)
	UnlockWorkspaceFn     func(id string) (*tfe.Workspace, error)
}

func (s WorkspaceService) CreateWorkspace(org string, opts *tfe.WorkspaceCreateOptions) (*tfe.Workspace, error) {
	return s.CreateWorkspaceFn(org, opts)
}

func (s WorkspaceService) UpdateWorkspace(name, org string, opts *tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error) {
	return s.UpdateWorkspaceFn(name, org, opts)
}

func (s WorkspaceService) UpdateWorkspaceByID(id string, opts *tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error) {
	return s.UpdateWorkspaceByIDFn(id, opts)
}

func (s WorkspaceService) GetWorkspace(name, org string) (*tfe.Workspace, error) {
	return s.GetWorkspaceFn(name, org)
}

func (s WorkspaceService) GetWorkspaceByID(id string) (*tfe.Workspace, error) {
	return s.GetWorkspaceByIDFn(id)
}

func (s WorkspaceService) ListWorkspaces(org string, opts ots.WorkspaceListOptions) (*ots.WorkspaceList, error) {
	return s.ListWorkspaceFn(org, opts)
}

func (s WorkspaceService) DeleteWorkspace(name, org string) error {
	return s.DeleteWorkspaceFn(name, org)
}

func (s WorkspaceService) DeleteWorkspaceByID(id string) error {
	return s.DeleteWorkspaceByIDFn(id)
}

func (s WorkspaceService) LockWorkspace(id string, opts ots.WorkspaceLockOptions) (*tfe.Workspace, error) {
	return s.LockWorkspaceFn(id, opts)
}

func (s WorkspaceService) UnlockWorkspace(id string) (*tfe.Workspace, error) {
	return s.UnlockWorkspaceFn(id)
}

func NewWorkspace(name, id, org string) *tfe.Workspace {
	return &tfe.Workspace{
		Actions: &tfe.WorkspaceActions{},
		ID:      id,
		Name:    name,
		Organization: &tfe.Organization{
			Name: org,
		},
		Permissions:     &tfe.WorkspacePermissions{},
		TriggerPrefixes: []string{},
		VCSRepo:         &tfe.VCSRepo{},
	}
}

func NewWorkspaceList(name, id, org string, opts ots.WorkspaceListOptions) *ots.WorkspaceList {
	return &ots.WorkspaceList{
		Items: []*tfe.Workspace{
			NewWorkspace(name, id, org),
		},
		Pagination: ots.NewPagination(opts.ListOptions, 1),
	}
}
