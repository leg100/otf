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
	return &FakeClient{
		fakeWorkspacesClient: &fakeWorkspacesClient{
			workspace: f.Workspace,
		},
	}, nil
}

type FakeClient struct {
	*fakeOrganizationsClient
	*fakeWorkspacesClient

	// TODO: stubbed until implemented
	otf.UserService
	otf.RunService
	otf.StateVersionService
	otf.ConfigurationVersionService
	otf.EventService
	otf.AgentTokenService
	otf.LatestRunService
}

type fakeOrganizationsClient struct {
	otf.OrganizationService
}

func (f *fakeOrganizationsClient) CreateOrganization(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	return otf.NewOrganization(otf.OrganizationCreateOptions{Name: otf.String(uuid.NewString())})
}

type fakeWorkspacesClient struct {
	otf.WorkspaceService
	workspace *otf.Workspace
}

func (f *fakeWorkspacesClient) GetWorkspace(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	return f.workspace, nil
}

func (f *fakeWorkspacesClient) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items:      []*otf.Workspace{f.workspace},
		Pagination: otf.NewPagination(otf.ListOptions{}, 1),
	}, nil
}

func (f *fakeWorkspacesClient) UpdateWorkspace(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	return f.workspace, nil
}

func (f *fakeWorkspacesClient) LockWorkspace(ctx context.Context, spec otf.WorkspaceSpec, _ otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	return f.workspace, nil
}

func (f *fakeWorkspacesClient) UnlockWorkspace(ctx context.Context, spec otf.WorkspaceSpec, _ otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	return f.workspace, nil
}
