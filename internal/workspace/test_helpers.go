package workspace

import (
	"context"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
)

type FakeService struct {
	Workspaces []*Workspace
	Policy     authz.WorkspacePolicy
}

func (f *FakeService) ListConnectedWorkspaces(ctx context.Context, vcsProviderID resource.ID, repoPath string) ([]*Workspace, error) {
	return f.Workspaces, nil
}

func (f *FakeService) Create(context.Context, CreateOptions) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) Update(_ context.Context, _ resource.ID, opts UpdateOptions) (*Workspace, error) {
	f.Workspaces[0].Update(opts)
	return f.Workspaces[0], nil
}

func (f *FakeService) List(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error) {
	return resource.NewPage(f.Workspaces, opts.PageOptions, nil), nil
}

func (f *FakeService) Get(context.Context, resource.ID) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) GetByName(context.Context, string, string) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) Delete(context.Context, resource.ID) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) Lock(context.Context, resource.ID, *resource.ID) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) Unlock(context.Context, resource.ID, *resource.ID, bool) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) ListTags(context.Context, string, ListTagsOptions) (*resource.Page[*Tag], error) {
	return nil, nil
}

func (f *FakeService) GetPolicy(context.Context, resource.ID) (authz.WorkspacePolicy, error) {
	return f.Policy, nil
}

func (f *FakeService) AddTags(ctx context.Context, workspaceID resource.ID, tags []TagSpec) error {
	return nil
}

func (f *FakeService) RemoveTags(ctx context.Context, workspaceID resource.ID, tags []TagSpec) error {
	return nil
}

func (f *FakeService) SetPermission(ctx context.Context, workspaceID, teamID resource.ID, role rbac.Role) error {
	return nil
}

func (f *FakeService) UnsetPermission(ctx context.Context, workspaceID, teamID resource.ID) error {
	return nil
}

type fakeVCSProviderService struct {
	providers []*vcsprovider.VCSProvider
	repos     []string
}

func (f *fakeVCSProviderService) Get(ctx context.Context, providerID resource.ID) (*vcsprovider.VCSProvider, error) {
	return f.providers[0], nil
}

func (f *fakeVCSProviderService) List(context.Context, string) ([]*vcsprovider.VCSProvider, error) {
	return f.providers, nil
}

func (f *fakeVCSProviderService) GetVCSClient(ctx context.Context, providerID resource.ID) (vcs.Client, error) {
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

func (f *fakeTeamService) List(context.Context, string) ([]*team.Team, error) {
	return f.teams, nil
}
