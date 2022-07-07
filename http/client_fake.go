package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf"
)

var _ ClientFactory = (*FakeClientFactory)(nil)

type FakeClientFactory struct {
	Workspace *otf.Workspace
}

func (f FakeClientFactory) NewClient() (Client, error) {
	return &FakeClient{workspaces: f.Workspace}, nil
}

type FakeClient struct {
	Client
	workspaces *otf.Workspace
}

func (f FakeClient) Organizations() otf.OrganizationService { return &FakeOrganizationsClient{} }

func (f FakeClient) Workspaces() otf.WorkspaceService {
	return &FakeWorkspacesClient{workspace: f.workspaces}
}

type FakeOrganizationsClient struct {
	otf.OrganizationService
}

func (f *FakeOrganizationsClient) Create(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	return otf.NewOrganization(otf.OrganizationCreateOptions{Name: otf.String(uuid.NewString())})
}

type FakeWorkspacesClient struct {
	otf.WorkspaceService
	workspace *otf.Workspace
}

func (f *FakeWorkspacesClient) Get(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	return f.workspace, nil
}

func (f *FakeWorkspacesClient) List(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items:      []*otf.Workspace{f.workspace},
		Pagination: otf.NewPagination(otf.ListOptions{}, 1),
	}, nil
}

func (f *FakeWorkspacesClient) Update(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	return f.workspace, nil
}

func (f *FakeWorkspacesClient) Lock(ctx context.Context, spec otf.WorkspaceSpec, _ otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	return f.workspace, nil
}

func (f *FakeWorkspacesClient) Unlock(ctx context.Context, spec otf.WorkspaceSpec, _ otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	return f.workspace, nil
}
