package http

import (
	"context"

	"github.com/leg100/otf"
)

var _ ClientFactory = (*FakeClientFactory)(nil)

type FakeClientFactory struct{}

func (f FakeClientFactory) NewClient() (Client, error) { return &FakeClient{}, nil }

type FakeClient struct {
	Client
}

func (f FakeClient) Organizations() otf.OrganizationService { return &FakeOrganizationsClient{} }

func (f FakeClient) Workspaces() otf.WorkspaceService { return &FakeWorkspacesClient{} }

type FakeOrganizationsClient struct {
	otf.OrganizationService
}

func (f *FakeOrganizationsClient) Create(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	return otf.NewTestOrganization(), nil
}

type FakeWorkspacesClient struct {
	otf.WorkspaceService
}

func (f *FakeWorkspacesClient) Get(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	return &otf.Workspace{
		ID: "ws-123",
	}, nil
}

func (f *FakeWorkspacesClient) List(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items: []*otf.Workspace{
			{
				ID: "ws-123",
			},
		},
	}, nil
}

func (f *FakeWorkspacesClient) Update(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	return &otf.Workspace{
		ID: "ws-123",
	}, nil
}

func (f *FakeWorkspacesClient) Lock(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	return &otf.Workspace{
		ID: "ws-123",
	}, nil
}

func (f *FakeWorkspacesClient) Unlock(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	return &otf.Workspace{
		ID: "ws-123",
	}, nil
}
