package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo(t *testing.T) {
	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})

	t.Run("connect workspace", func(t *testing.T) {
		svc := setup(t, "test/dummy")

		org := svc.createOrganization(t, ctx)
		ws := svc.createWorkspace(t, ctx, org)
		vcsprov := svc.createVCSProvider(t, ctx, org)

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

	t.Run("create workspace with connection", func(t *testing.T) {
		svc := setup(t, "test/dummy")

		org := svc.createOrganization(t, ctx)
		vcsprov := svc.createVCSProvider(t, ctx, org)
		ws, err := svc.CreateWorkspace(ctx, workspace.CreateOptions{
			Name:         otf.String(uuid.NewString()),
			Organization: &org.Name,
			ConnectOptions: &workspace.ConnectOptions{
				RepoPath:      "test/dummy",
				VCSProviderID: vcsprov.ID,
			},
		})
		require.NoError(t, err)

		// webhook should be registered with github
		require.True(t, svc.githubServer.HasWebhook())

		t.Run("delete workspace connection", func(t *testing.T) {
			err := svc.Disconnect(ctx, repo.DisconnectOptions{
				ConnectionType: repo.WorkspaceConnection,
				ResourceID:     ws.ID,
			})
			require.NoError(t, err)
		})

		// webhook should now have been deleted from github
		require.False(t, svc.githubServer.HasWebhook())
	})

	t.Run("create module with connection", func(t *testing.T) {
		svc := setup(t, "leg100/terraform-aws-stuff")

		org := svc.createOrganization(t, ctx)
		vcsprov := svc.createVCSProvider(t, ctx, org)
		mod := svc.createModule(t, ctx, org)

		mod, err := svc.PublishModule(ctx, module.PublishOptions{
			VCSProviderID: vcsprov.ID,
			Repo:          module.Repo("leg100/terraform-aws-stuff"),
		})
		require.NoError(t, err)

		// webhook should be registered with github
		require.True(t, svc.githubServer.HasWebhook())

		t.Run("delete module", func(t *testing.T) {
			_, err := svc.DeleteModule(ctx, mod.ID)
			require.NoError(t, err)
		})

		// webhook should now have been deleted from github
		require.False(t, svc.githubServer.HasWebhook())
	})

	t.Run("create multiple connections", func(t *testing.T) {
		svc := setup(t, "test/dummy")

		org := svc.createOrganization(t, ctx)
		vcsprov := svc.createVCSProvider(t, ctx, org)

		mod1 := svc.createModule(t, ctx, org)
		_, err := svc.Connect(ctx, repo.ConnectOptions{
			ConnectionType: repo.ModuleConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     mod1.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)

		mod2 := svc.createModule(t, ctx, org)
		_, err = svc.Connect(ctx, repo.ConnectOptions{
			ConnectionType: repo.ModuleConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     mod2.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)

		ws1 := svc.createWorkspace(t, ctx, org)
		_, err = svc.Connect(ctx, repo.ConnectOptions{
			ConnectionType: repo.WorkspaceConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     ws1.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)

		ws2 := svc.createWorkspace(t, ctx, org)
		_, err = svc.Connect(ctx, repo.ConnectOptions{
			ConnectionType: repo.WorkspaceConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     ws2.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)

		// webhook should be registered with github
		require.True(t, svc.githubServer.HasWebhook())

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
			require.False(t, svc.githubServer.HasWebhook())
		})
	})
}
