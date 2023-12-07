package workspace

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
)

type FakeService struct {
	Workspaces []*Workspace
	Policy     internal.WorkspacePolicy
}

func (f *FakeService) ListConnectedWorkspaces(ctx context.Context, vcsProviderID, repoPath string) ([]*Workspace, error) {
	return f.Workspaces, nil
}

func (f *FakeService) CreateWorkspace(context.Context, CreateOptions) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) UpdateWorkspace(_ context.Context, _ string, opts UpdateOptions) (*Workspace, error) {
	f.Workspaces[0].Update(opts)
	return f.Workspaces[0], nil
}

func (f *FakeService) ListWorkspaces(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error) {
	return resource.NewPage(f.Workspaces, opts.PageOptions, nil), nil
}

func (f *FakeService) GetWorkspace(context.Context, string) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) GetWorkspaceByName(context.Context, string, string) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) DeleteWorkspace(context.Context, string) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) LockWorkspace(context.Context, string, *string) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) UnlockWorkspace(context.Context, string, *string, bool) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) ListTags(context.Context, string, ListTagsOptions) (*resource.Page[*Tag], error) {
	return nil, nil
}

func (f *FakeService) GetPolicy(context.Context, string) (internal.WorkspacePolicy, error) {
	return f.Policy, nil
}

func (f *FakeService) AddTags(ctx context.Context, workspaceID string, tags []TagSpec) error {
	return nil
}

func (f *FakeService) RemoveTags(ctx context.Context, workspaceID string, tags []TagSpec) error {
	return nil
}

func (f *FakeService) SetPermission(ctx context.Context, workspaceID, teamID string, role rbac.Role) error {
	return nil
}

func (f *FakeService) UnsetPermission(ctx context.Context, workspaceID, teamID string) error {
	return nil
}

type fakeVCSProviderService struct {
	providers []*vcsprovider.VCSProvider
	repos     []string
}

func (f *fakeVCSProviderService) GetVCSProvider(ctx context.Context, providerID string) (*vcsprovider.VCSProvider, error) {
	return f.providers[0], nil
}

func (f *fakeVCSProviderService) ListVCSProviders(context.Context, string) ([]*vcsprovider.VCSProvider, error) {
	return f.providers, nil
}

func (f *fakeVCSProviderService) GetVCSClient(ctx context.Context, providerID string) (vcs.Client, error) {
	return &fakeVCSClient{repos: f.repos}, nil
}

type fakeVCSClient struct {
	repos []string

	vcs.Client
}

func (f *fakeVCSClient) ListRepositories(ctx context.Context, opts vcs.ListRepositoriesOptions) ([]string, error) {
	return f.repos, nil
}

type fakeTeamService struct {
	teams []*team.Team
}

func (f *fakeTeamService) ListTeams(context.Context, string) ([]*team.Team, error) {
	return f.teams, nil
}
