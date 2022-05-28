package http

import (
	"context"

	"github.com/google/uuid"
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
	return otf.NewOrganization(otf.OrganizationCreateOptions{Name: otf.String(uuid.NewString())})
}

type FakeWorkspacesClient struct {
	otf.WorkspaceService
}

func (f *FakeWorkspacesClient) Get(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	return (&otf.WorkspaceFactory{}).NewWorkspace(context.Background(), otf.WorkspaceCreateOptions{
		Name:           uuid.NewString(),
		OrganizationID: otf.String("org-123"),
	})
}

func (f *FakeWorkspacesClient) List(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	ws, err := (&otf.WorkspaceFactory{}).NewWorkspace(context.Background(), otf.WorkspaceCreateOptions{
		Name:           uuid.NewString(),
		OrganizationID: otf.String("org-123"),
	})
	if err != nil {
		return nil, err
	}
	return &otf.WorkspaceList{
		Items: []*otf.Workspace{ws},
	}, nil
}

func (f *FakeWorkspacesClient) Update(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	return (&otf.WorkspaceFactory{}).NewWorkspace(context.Background(), otf.WorkspaceCreateOptions{
		Name:           uuid.NewString(),
		OrganizationID: otf.String("org-123"),
	})
}

func (f *FakeWorkspacesClient) Lock(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	return (&otf.WorkspaceFactory{}).NewWorkspace(context.Background(), otf.WorkspaceCreateOptions{
		Name:           uuid.NewString(),
		OrganizationID: otf.String("org-123"),
	})
}

func (f *FakeWorkspacesClient) Unlock(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	return (&otf.WorkspaceFactory{}).NewWorkspace(context.Background(), otf.WorkspaceCreateOptions{
		Name:           uuid.NewString(),
		OrganizationID: otf.String("org-123"),
	})
}
