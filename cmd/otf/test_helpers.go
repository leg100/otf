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

func withFakeWorkspaces(workspaces ...*otf.Workspace) fakeOption {
	return func(c *fakeClient) {
		c.workspaces = workspaces
	}
}

func withFakeRun(run *otf.Run) fakeOption {
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
	workspaces []*otf.Workspace
	run        *otf.Run
	tarball    []byte
	otf.Client
}

func (f *fakeClient) CreateOrganization(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	return otf.NewOrganization(opts)
}

func (f *fakeClient) GetWorkspace(context.Context, string) (*otf.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeClient) GetWorkspaceByName(context.Context, string, string) (*otf.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeClient) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items:      f.workspaces,
		Pagination: otf.NewPagination(otf.ListOptions{}, len(f.workspaces)),
	}, nil
}

func (f *fakeClient) ListVariables(ctx context.Context, workspaceID string) ([]otf.Variable, error) {
	return nil, nil
}

func (f *fakeClient) UpdateWorkspace(ctx context.Context, workspaceID string, opts otf.UpdateWorkspaceOptions) (*otf.Workspace, error) {
	f.workspaces[0].Update(opts)
	return f.workspaces[0], nil
}

func (f *fakeClient) LockWorkspace(context.Context, string, otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeClient) UnlockWorkspace(context.Context, string, otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeClient) GetRun(context.Context, string) (*otf.Run, error) {
	return f.run, nil
}

func (f *fakeClient) DownloadConfig(context.Context, string) ([]byte, error) {
	return f.tarball, nil
}

func (f *fakeClient) CreateAgentToken(ctx context.Context, opts otf.CreateAgentTokenOptions) (*otf.AgentToken, error) {
	return otf.NewAgentToken(opts)
}
