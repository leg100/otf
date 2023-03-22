package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/github"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/vcsprovider"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/require"
)

type service struct {
	vcsprovider.VCSProviderService
	module.ModuleService
	workspace.WorkspaceService
	repo.RepoService
	organization.OrganizationService
	cloud.Service

	githubServer *github.TestServer
}

func setup(t *testing.T, repo string) *service {
	db := sql.NewTestDB(t)
	cloudService, githubServer := newCloudService(t, repo)
	vcsproviderService := newVCSProviderService(t, db, cloudService)
	repoService := newRepoService(t, db, cloudService, vcsproviderService)

	return &service{
		OrganizationService: newOrganizationService(t, db),
		Service:             cloudService,
		WorkspaceService:    newWorkspaceService(t, db, repoService),
		VCSProviderService:  vcsproviderService,
		RepoService:         repoService,
		githubServer:        githubServer,
		ModuleService:       newModuleService(t, db, repoService, vcsproviderService),
	}
}

func (s *service) createOrganization(t *testing.T, ctx context.Context) *organization.Organization {
	org, err := s.CreateOrganization(ctx, organization.OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)
	return org
}

func (s *service) createWorkspace(t *testing.T, ctx context.Context, org *organization.Organization, opts *workspace.CreateOptions) *workspace.Workspace {
	if org != nil {
		ws, err := s.CreateWorkspace(ctx, workspace.CreateOptions{
			Name:         otf.String(uuid.NewString()),
			Organization: &org.Name,
		})
		require.NoError(t, err)
		return ws
	} else {
		ws, err := s.CreateWorkspace(ctx, *opts)
		require.NoError(t, err)
		return ws
	}
}

func (s *service) createVCSProvider(t *testing.T, ctx context.Context, org *organization.Organization) *vcsprovider.VCSProvider {
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

func (s *service) createModule(t *testing.T, ctx context.Context, org *organization.Organization) *module.Module {
	module, err := s.CreateModule(ctx, module.CreateOptions{
		Name:         uuid.NewString(),
		Provider:     uuid.NewString(),
		Organization: org.Name,
	})
	require.NoError(t, err)
	return module
}

func (s *service) hasWebhook() bool {
	return s.githubServer.HookEndpoint != nil && s.githubServer.HookSecret != nil
}
