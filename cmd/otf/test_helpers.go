package main

import (
	"context"

	"github.com/leg100/otf"
)

func fakeApp(opts ...fakeOption) *application {
	client := fakeClient{}
	for _, fn := range opts {
		fn(&client)
	}
	return &application{&client}
}

type fakeOption func(*fakeClient)

func withFakeWorkspaces(workspaces ...*workspace.Workspace) fakeOption {
	return func(c *fakeClient) {
		c.workspaces = workspaces
	}
}

func withFakeRun(run *run.Run) fakeOption {
	return func(c *fakeClient) {
		c.run = run
	}
}

func withFakeTarball(tarball []byte) fakeOption {
	return func(c *fakeClient) {
		c.tarball = tarball
	}
}

type fakeClient struct {
	workspaces []*workspace.Workspace
	run        *run.Run
	tarball    []byte
	otf.Client
}

func (f *fakeClient) CreateOrganization(ctx context.Context, opts organization.OrganizationCreateOptions) (*organization.Organization, error) {
	return otf.NewOrganization(opts)
}

func (f *fakeClient) GetWorkspace(context.Context, string) (*workspace.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeClient) GetWorkspaceByName(context.Context, string, string) (*workspace.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeClient) ListWorkspaces(ctx context.Context, opts workspace.WorkspaceListOptions) (*workspace.WorkspaceList, error) {
	return &workspace.WorkspaceList{
		Items:      f.workspaces,
		Pagination: otf.NewPagination(otf.ListOptions{}, len(f.workspaces)),
	}, nil
}

func (f *fakeClient) ListVariables(ctx context.Context, workspaceID string) ([]otf.Variable, error) {
	return nil, nil
}

func (f *fakeClient) UpdateWorkspace(ctx context.Context, workspaceID string, opts otf.UpdateWorkspaceOptions) (*workspace.Workspace, error) {
	f.workspaces[0].Update(opts)
	return f.workspaces[0], nil
}

func (f *fakeClient) LockWorkspace(context.Context, string, workspace.WorkspaceLockOptions) (*workspace.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeClient) UnlockWorkspace(context.Context, string, workspace.WorkspaceUnlockOptions) (*workspace.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeClient) GetRun(context.Context, string) (*run.Run, error) {
	return f.run, nil
}

func (f *fakeClient) DownloadConfig(context.Context, string) ([]byte, error) {
	return f.tarball, nil
}

func (f *fakeClient) CreateAgentToken(ctx context.Context, opts otf.CreateAgentTokenOptions) (*otf.AgentToken, error) {
	return otf.NewAgentToken(opts)
}
