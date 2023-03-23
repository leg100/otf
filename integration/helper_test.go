package integration

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/github"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/services"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/vcsprovider"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/require"
)

type testServices struct {
	*services.Services

	githubServer *github.TestServer
}

func setup(t *testing.T, repo string) *testServices {
	db := sql.NewTestDB(t)
	cfg := services.NewDefaultConfig()

	// Use stub github server
	githubServer, githubCfg := github.NewTestServer(t, github.WithRepo(repo))
	cfg.Github.Config = githubCfg

	svcs, _, err := services.New(logr.Discard(), db, cfg)
	require.NoError(t, err)
	return &testServices{
		Services:     svcs,
		githubServer: githubServer,
	}
}

func (s *testServices) createOrganization(t *testing.T, ctx context.Context) *organization.Organization {
	org, err := s.CreateOrganization(ctx, organization.OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)
	return org
}

func (s *testServices) createWorkspace(t *testing.T, ctx context.Context, org *organization.Organization) *workspace.Workspace {
	ws, err := s.CreateWorkspace(ctx, workspace.CreateOptions{
		Name:         otf.String(uuid.NewString()),
		Organization: &org.Name,
	})
	require.NoError(t, err)
	return ws
}

func (s *testServices) createVCSProvider(t *testing.T, ctx context.Context, org *organization.Organization) *vcsprovider.VCSProvider {
	provider, err := s.CreateVCSProvider(ctx, vcsprovider.CreateOptions{
		Organization: org.Name,
		// tests require a legitimate cloud name to avoid invalid foreign
		// key error upon insert/update
		Cloud: "github",
		Name:  uuid.NewString(),
		Token: uuid.NewString(),
	})
	require.NoError(t, err)
	return provider
}

func (s *testServices) createModule(t *testing.T, ctx context.Context, org *organization.Organization) *module.Module {
	module, err := s.CreateModule(ctx, module.CreateOptions{
		Name:         uuid.NewString(),
		Provider:     uuid.NewString(),
		Organization: org.Name,
	})
	require.NoError(t, err)
	return module
}
