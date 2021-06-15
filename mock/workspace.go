package mock

import (
	"github.com/hashicorp/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.WorkspaceService = (*WorkspaceService)(nil)

type WorkspaceService struct {
	CreateWorkspaceFn     func(org string, opts *ots.WorkspaceCreateOptions) (*ots.Workspace, error)
	UpdateWorkspaceFn     func(name, org string, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error)
	UpdateWorkspaceByIDFn func(id string, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error)
	GetWorkspaceFn        func(name, org string) (*ots.Workspace, error)
	GetWorkspaceByIDFn    func(id string) (*ots.Workspace, error)
	ListWorkspaceFn       func(org string, opts ots.WorkspaceListOptions) (*ots.WorkspaceList, error)
	DeleteWorkspaceFn     func(name, org string) error
	DeleteWorkspaceByIDFn func(id string) error
}

func (s WorkspaceService) CreateWorkspace(org string, opts *ots.WorkspaceCreateOptions) (*ots.Workspace, error) {
	return s.CreateWorkspaceFn(org, opts)
}

func (s WorkspaceService) UpdateWorkspace(name, org string, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error) {
	return s.UpdateWorkspaceFn(name, org, opts)
}

func (s WorkspaceService) UpdateWorkspaceByID(id string, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error) {
	return s.UpdateWorkspaceByIDFn(id, opts)
}

func (s WorkspaceService) GetWorkspace(name, org string) (*ots.Workspace, error) {
	return s.GetWorkspaceFn(name, org)
}

func (s WorkspaceService) GetWorkspaceByID(id string) (*ots.Workspace, error) {
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

func NewWorkspace(name, id, org string) *ots.Workspace {
	return &ots.Workspace{
		Actions: &ots.WorkspaceActions{},
		ID:      id,
		Name:    name,
		Organization: &ots.Organization{
			Name: org,
		},
		Permissions:     &ots.WorkspacePermissions{},
		TriggerPrefixes: []string{},
		VCSRepo:         &ots.VCSRepo{},
	}
}

func NewWorkspaceList(name, id, org string, opts ots.WorkspaceListOptions) *ots.WorkspaceList {
	return &ots.WorkspaceList{
		Items: []*ots.Workspace{
			NewWorkspace(name, id, org),
		},
	}
}
