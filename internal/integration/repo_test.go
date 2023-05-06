package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/github"
	"github.com/leg100/otf/repo"
	"github.com/stretchr/testify/require"
)

func TestRepo(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &auth.SiteAdmin)

	t.Run("create multiple connections", func(t *testing.T) {
		svc := setup(t, nil, github.WithRepo("test/dummy"))

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
		require.True(t, svc.HasWebhook())

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
			require.False(t, svc.HasWebhook())
		})
	})
}
