package cli

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/client"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/orgcreator"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
)

type (
	fakeClient struct {
		user             *auth.User
		team             *auth.Team
		workspaces       []*workspace.Workspace
		run              *run.Run
		stateVersion     *state.Version
		stateVersionList *state.VersionList
		state            []byte
		agentToken       []byte
		tarball          []byte
		client.Client
	}

	fakeOption func(*fakeClient)
)

func fakeApp(opts ...fakeOption) *CLI {
	client := fakeClient{}
	for _, fn := range opts {
		fn(&client)
	}
	return &CLI{&client, ""}
}

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

func withStateVersion(sv *state.Version) fakeOption {
	return func(c *fakeClient) {
		c.stateVersion = sv
	}
}

func withStateVersionList(svl *state.VersionList) fakeOption {
	return func(c *fakeClient) {
		c.stateVersionList = svl
	}
}

func withState(state []byte) fakeOption {
	return func(c *fakeClient) {
		c.state = state
	}
}

func withAgentToken(token []byte) fakeOption {
	return func(c *fakeClient) {
		c.agentToken = token
	}
}

func withTarball(tarball []byte) fakeOption {
	return func(c *fakeClient) {
		c.tarball = tarball
	}
}

func (f *fakeClient) CreateOrganization(ctx context.Context, opts orgcreator.OrganizationCreateOptions) (*organization.Organization, error) {
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

func (f *fakeClient) AddTeamMembership(context.Context, auth.TeamMembershipOptions) error {
	return nil
}

func (f *fakeClient) RemoveTeamMembership(context.Context, auth.TeamMembershipOptions) error {
	return nil
}

func (f *fakeClient) CreateTeam(context.Context, string, auth.CreateTeamOptions) (*auth.Team, error) {
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
		Pagination: internal.NewPagination(internal.ListOptions{}, len(f.workspaces)),
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

func (f *fakeClient) CreateAgentToken(ctx context.Context, opts tokens.CreateAgentTokenOptions) ([]byte, error) {
	return f.agentToken, nil
}

func (f *fakeClient) ListStateVersions(context.Context, string, internal.ListOptions) (*state.VersionList, error) {
	return f.stateVersionList, nil
}

func (f *fakeClient) GetCurrentStateVersion(ctx context.Context, workspaceID string) (*state.Version, error) {
	if f.stateVersion == nil {
		return nil, internal.ErrResourceNotFound
	}
	return f.stateVersion, nil
}

func (f *fakeClient) DeleteStateVersion(ctx context.Context, svID string) error {
	return nil
}

func (f *fakeClient) RollbackStateVersion(ctx context.Context, svID string) (*state.Version, error) {
	return f.stateVersion, nil
}

func (f *fakeClient) DownloadState(ctx context.Context, svID string) ([]byte, error) {
	return f.state, nil
}
