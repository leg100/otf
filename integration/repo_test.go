package integration

import (
	"context"
	"net/url"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/github"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/vcsprovider"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)

	t.Run("create workspace connection", func(t *testing.T) {
		cloudService, _ := testCloudService(t, "test/dummy")
		svc := testRepoService(t, db, cloudService)
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		vcsprov := vcsprovider.CreateTestVCSProvider(t, db, org, vcsprovider.WithCloudService(cloudService))

		got, err := svc.Connect(ctx, repo.ConnectOptions{
			ConnectionType: repo.WorkspaceConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     ws.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)
		want := &repo.Connection{VCSProviderID: vcsprov.ID, Repo: "test/dummy"}
		assert.Equal(t, want, got)

		t.Run("delete workspace connection", func(t *testing.T) {
			err := svc.Disconnect(ctx, repo.DisconnectOptions{
				ConnectionType: repo.WorkspaceConnection,
				ResourceID:     ws.ID,
			})
			require.NoError(t, err)
		})
	})

	t.Run("create module connection", func(t *testing.T) {
		cloudService, _ := testCloudService(t, "test/dummy")
		svc := testRepoService(t, db, cloudService)
		org := organization.CreateTestOrganization(t, db)
		mod := module.CreateTestModule(t, db, org)
		vcsprov := vcsprovider.CreateTestVCSProvider(t, db, org, vcsprovider.WithCloudService(cloudService))

		got, err := svc.Connect(ctx, repo.ConnectOptions{
			ConnectionType: repo.ModuleConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     mod.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)
		want := &repo.Connection{VCSProviderID: vcsprov.ID, Repo: "test/dummy"}
		assert.Equal(t, want, got)

		t.Run("delete module connection", func(t *testing.T) {
			err := svc.Disconnect(ctx, repo.DisconnectOptions{
				ConnectionType: repo.ModuleConnection,
				ResourceID:     mod.ID,
			})
			require.NoError(t, err)
		})
	})

	t.Run("create multiple connections", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		cloudService, githubServer := testCloudService(t, "test/dummy")
		vcsprov := vcsprovider.CreateTestVCSProvider(t, db, org, vcsprovider.WithCloudService(cloudService))
		svc := testRepoService(t, db, cloudService)

		mod1 := module.CreateTestModule(t, db, org)
		_, err := svc.Connect(ctx, repo.ConnectOptions{
			ConnectionType: repo.ModuleConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     mod1.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)

		mod2 := module.CreateTestModule(t, db, org)
		_, err = svc.Connect(ctx, repo.ConnectOptions{
			ConnectionType: repo.ModuleConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     mod2.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)

		ws1 := workspace.CreateTestWorkspace(t, db, org.Name)
		_, err = svc.Connect(ctx, repo.ConnectOptions{
			ConnectionType: repo.WorkspaceConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     ws1.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)

		ws2 := workspace.CreateTestWorkspace(t, db, org.Name)
		_, err = svc.Connect(ctx, repo.ConnectOptions{
			ConnectionType: repo.WorkspaceConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     ws2.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)

		// webhook should be registered with github
		require.NotNil(t, githubServer.HookEndpoint)
		require.NotNil(t, githubServer.HookSecret)

		t.Run("delete multiple connections", func(t *testing.T) {
			err = svc.Disconnect(ctx, repo.DisconnectOptions{
				ConnectionType: repo.WorkspaceConnection,
				ResourceID:     ws2.ID,
			})
			require.NoError(t, err)

			err = svc.Disconnect(ctx, repo.DisconnectOptions{
				ConnectionType: repo.WorkspaceConnection,
				ResourceID:     ws1.ID,
			})
			require.NoError(t, err)

			err := svc.Disconnect(ctx, repo.DisconnectOptions{
				ConnectionType: repo.ModuleConnection,
				ResourceID:     mod2.ID,
			})
			require.NoError(t, err)

			err = svc.Disconnect(ctx, repo.DisconnectOptions{
				ConnectionType: repo.ModuleConnection,
				ResourceID:     mod1.ID,
			})
			require.NoError(t, err)

			// webhook should now have been deleted from github
			require.Nil(t, githubServer.HookEndpoint)
			require.Nil(t, githubServer.HookSecret)
		})
	})
}

func createTestConnection(t *testing.T, db otf.DB, repoPath string) *repo.Connection {
	cloudService, _ := testCloudService(t, repoPath)
	svc := testRepoService(t, db, cloudService)
	org := organization.CreateTestOrganization(t, db)
	mod := module.CreateTestModule(t, db, org)
	vcsprov := vcsprovider.CreateTestVCSProvider(t, db, org, vcsprovider.WithCloudService(cloudService))

	connection, err := svc.Connect(context.Background(), repo.ConnectOptions{
		ConnectionType: repo.ModuleConnection,
		VCSProviderID:  vcsprov.ID,
		ResourceID:     mod.ID,
		RepoPath:       repoPath,
	})
	require.NoError(t, err)
	return connection
}

func testRepoService(t *testing.T, db otf.DB, cloudService cloud.Service) *repo.Service {
	return repo.NewService(repo.Options{
		Logger:             logr.Discard(),
		DB:                 db,
		CloudService:       cloudService,
		HostnameService:    hostnameService{"fake-host.org"},
		VCSProviderService: vcsprovider.NewTestService(t, db, vcsprovider.WithCloudService(cloudService)),
	})
}

func testCloudService(t *testing.T, repoPath string) (cloud.Service, *github.TestServer) {
	srv := github.NewTestServer(t, github.WithRepo(repoPath))
	githubURL, err := url.Parse(srv.URL)
	require.NoError(t, err)
	githubCloudConfig := cloud.Config{
		Name:                "github",
		Hostname:            githubURL.Host,
		Cloud:               &github.Cloud{},
		SkipTLSVerification: true,
	}
	svc, err := inmem.NewCloudService(githubCloudConfig)
	require.NoError(t, err)
	return svc, srv

}
