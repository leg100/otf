package main

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/client"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/workspace"
)

func fakeApp(opts ...fakeOption) *application {
	client := fakeClient{}
	for _, fn := range opts {
		fn(&client)
	}
	return &application{&client, ""}
}

type fakeOption func(*fakeClient)

func withUser(user *auth.User) fakeOption {
	return func(c *fakeClient) {
		c.user = user
	}
}

func withTeam(team *auth.Team) fakeOption {
	return func(c *fakeClient) {
		c.team = team
	}
}

func withWorkspaces(workspaces ...*workspace.Workspace) fakeOption {
	return func(c *fakeClient) {
		c.workspaces = workspaces
	}
}

func withRun(run *run.Run) fakeOption {
	return func(c *fakeClient) {
		c.run = run
	}
}

func withAgentToken(at *auth.AgentToken) fakeOption {
	return func(c *fakeClient) {
		c.agentToken = at
	}
}

func withTarball(tarball []byte) fakeOption {
	return func(c *fakeClient) {
		c.tarball = tarball
	}
}

type fakeClient struct {
	user       *auth.User
	team       *auth.Team
	workspaces []*workspace.Workspace
	run        *run.Run
	agentToken *auth.AgentToken
	tarball    []byte
	client.Client
}

func (f *fakeClient) CreateOrganization(ctx context.Context, opts organization.OrganizationCreateOptions) (*organization.Organization, error) {
	return &organization.Organization{Name: *opts.Name}, nil
}

func (f *fakeClient) DeleteOrganization(context.Context, string) error {
	return nil
}

func (f *fakeClient) CreateUser(context.Context, string, ...auth.NewUserOption) (*auth.User, error) {
	return f.user, nil
}

func (f *fakeClient) DeleteUser(context.Context, string) error {
	return nil
}

func (f *fakeClient) AddTeamMembership(context.Context, string, string) error {
	return nil
}

func (f *fakeClient) RemoveTeamMembership(context.Context, string, string) error {
	return nil
}

func (f *fakeClient) CreateTeam(context.Context, auth.NewTeamOptions) (*auth.Team, error) {
	return f.team, nil
}

func (f *fakeClient) GetTeam(context.Context, string, string) (*auth.Team, error) {
	return f.team, nil
}

func (f *fakeClient) DeleteTeam(context.Context, string) error {
	return nil
}

func (f *fakeClient) GetWorkspace(context.Context, string) (*workspace.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeClient) GetWorkspaceByName(context.Context, string, string) (*workspace.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeClient) ListWorkspaces(ctx context.Context, opts workspace.ListOptions) (*workspace.WorkspaceList, error) {
	return &workspace.WorkspaceList{
		Items:      f.workspaces,
		Pagination: otf.NewPagination(otf.ListOptions{}, len(f.workspaces)),
	}, nil
}

func (f *fakeClient) ListVariables(ctx context.Context, workspaceID string) ([]*variable.Variable, error) {
	return nil, nil
}

func (f *fakeClient) UpdateWorkspace(ctx context.Context, workspaceID string, opts workspace.UpdateOptions) (*workspace.Workspace, error) {
	f.workspaces[0].Update(opts)
	return f.workspaces[0], nil
}

func (f *fakeClient) LockWorkspace(context.Context, string, *string) (*workspace.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeClient) UnlockWorkspace(context.Context, string, *string, bool) (*workspace.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeClient) GetRun(context.Context, string) (*run.Run, error) {
	return f.run, nil
}

func (f *fakeClient) DownloadConfig(context.Context, string) ([]byte, error) {
	return f.tarball, nil
}

func (f *fakeClient) CreateAgentToken(ctx context.Context, opts auth.CreateAgentTokenOptions) (*auth.AgentToken, error) {
	return f.agentToken, nil
}
