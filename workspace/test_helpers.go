package workspace

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/policy"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/vcsprovider"
	"github.com/stretchr/testify/require"
)

type (
	fakeWebService struct {
		workspaces []*Workspace
		providers  []*vcsprovider.VCSProvider
		repos      []string
		policy     otf.WorkspacePolicy
		teams      []*auth.Team

		Service

		auth.TeamService
		VCSProviderService
		policy.PolicyService
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

func withPolicy(policy otf.WorkspacePolicy) fakeWebServiceOption {
	return func(svc *fakeWebService) {
		svc.policy = policy
	}
}

func withTeams(teams ...*auth.Team) fakeWebServiceOption {
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
		PolicyService:      &svc,
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

func (f *fakeWebService) GetPolicy(context.Context, string) (otf.WorkspacePolicy, error) {
	return f.policy, nil
}

func (f *fakeWebService) ListTeams(context.Context, string) ([]*auth.Team, error) {
	return f.teams, nil
}

func (f *fakeWebService) GetVCSClient(ctx context.Context, providerID string) (cloud.Client, error) {
	return &fakeWebCloudClient{repos: f.repos}, nil
}

func (f *fakeWebService) CreateWorkspace(context.Context, CreateOptions) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) UpdateWorkspace(context.Context, string, UpdateOptions) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) ListWorkspaces(ctx context.Context, opts ListOptions) (*WorkspaceList, error) {
	return &WorkspaceList{
		Items:      f.workspaces,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.workspaces)),
	}, nil
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

func (f *fakeWebService) connect(context.Context, string, ConnectOptions) (*repo.Connection, error) {
	return nil, nil
}

func (f *fakeWebService) disconnect(context.Context, string) error {
	return nil
}

type fakeWebCloudClient struct {
	repos []string

	cloud.Client
}

func (f *fakeWebCloudClient) ListRepositories(ctx context.Context, opts cloud.ListRepositoriesOptions) ([]string, error) {
	return f.repos, nil
}
