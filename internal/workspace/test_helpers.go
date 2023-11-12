package workspace

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/stretchr/testify/require"
)

type (
	fakeWebService struct {
		workspaces []*Workspace
		providers  []*vcsprovider.VCSProvider
		repos      []string
		policy     internal.WorkspacePolicy
		teams      []*team.Team

		Service

		team.TeamService
		VCSProviderService
	}

	fakeWebServiceOption func(*fakeWebService)
)

func withWorkspaces(workspaces ...*Workspace) fakeWebServiceOption {
	return func(svc *fakeWebService) {
		svc.workspaces = workspaces
	}
}

func withVCSProviders(providers ...*vcsprovider.VCSProvider) fakeWebServiceOption {
	return func(svc *fakeWebService) {
		svc.providers = providers
	}
}

func withRepos(repos ...string) fakeWebServiceOption {
	return func(svc *fakeWebService) {
		svc.repos = repos
	}
}

func withPolicy(policy internal.WorkspacePolicy) fakeWebServiceOption {
	return func(svc *fakeWebService) {
		svc.policy = policy
	}
}

func withTeams(teams ...*team.Team) fakeWebServiceOption {
	return func(svc *fakeWebService) {
		svc.teams = teams
	}
}

func fakeWebHandlers(t *testing.T, opts ...fakeWebServiceOption) *webHandlers {
	renderer, err := html.NewRenderer(false)
	require.NoError(t, err)

	var svc fakeWebService
	for _, fn := range opts {
		fn(&svc)
	}

	return &webHandlers{
		Renderer:           renderer,
		TeamService:        &svc,
		VCSProviderService: &svc,
		svc:                &svc,
	}
}

func (f *fakeWebService) GetVCSProvider(ctx context.Context, providerID string) (*vcsprovider.VCSProvider, error) {
	return f.providers[0], nil
}

func (f *fakeWebService) ListVCSProviders(context.Context, string) ([]*vcsprovider.VCSProvider, error) {
	return f.providers, nil
}

func (f *fakeWebService) UploadConfig(context.Context, string, []byte) error {
	return nil
}

func (f *fakeWebService) GetPolicy(context.Context, string) (internal.WorkspacePolicy, error) {
	return f.policy, nil
}

func (f *fakeWebService) ListTeams(context.Context, string) ([]*team.Team, error) {
	return f.teams, nil
}

func (f *fakeWebService) GetVCSClient(ctx context.Context, providerID string) (vcs.Client, error) {
	return &fakeWebCloudClient{repos: f.repos}, nil
}

func (f *fakeWebService) CreateWorkspace(context.Context, CreateOptions) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) UpdateWorkspace(context.Context, string, UpdateOptions) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) ListWorkspaces(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error) {
	return resource.NewPage(f.workspaces, opts.PageOptions, nil), nil
}

func (f *fakeWebService) GetWorkspace(context.Context, string) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) GetWorkspaceByName(context.Context, string, string) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) DeleteWorkspace(context.Context, string) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) LockWorkspace(context.Context, string, *string) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) UnlockWorkspace(context.Context, string, *string, bool) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) ListTags(context.Context, string, ListTagsOptions) (*resource.Page[*Tag], error) {
	return nil, nil
}

type fakeWebCloudClient struct {
	repos []string

	vcs.Client
}

func (f *fakeWebCloudClient) ListRepositories(ctx context.Context, opts vcs.ListRepositoriesOptions) ([]string, error) {
	return f.repos, nil
}
