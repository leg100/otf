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
	return &otf.Organization{
		Name: *opts.Name,
	}, nil
}

type FakeWorkspacesClient struct {
	otf.WorkspaceService
}

func (f *FakeWorkspacesClient) Get(ctx context.Context, spec otf.WorkspaceSpecifier) (*otf.Workspace, error) {
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

func (f *FakeWorkspacesClient) Update(ctx context.Context, spec otf.WorkspaceSpecifier, opts otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	return &otf.Workspace{
		ID: "ws-123",
	}, nil
}

func (f *FakeWorkspacesClient) Lock(ctx context.Context, id string, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	return &otf.Workspace{
		ID: "ws-123",
	}, nil
}

func (f *FakeWorkspacesClient) Unlock(ctx context.Context, id string) (*otf.Workspace, error) {
	return &otf.Workspace{
		ID: "ws-123",
	}, nil
}
