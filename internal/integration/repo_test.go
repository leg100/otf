package integration

import (
	"testing"

	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/github"
	"github.com/stretchr/testify/require"
)

func TestRepo(t *testing.T) {
	integrationTest(t)

	t.Run("create multiple connections", func(t *testing.T) {
		svc, org, ctx := setup(t, nil, github.WithRepo("test/dummy"))

		vcsprov := svc.createVCSProvider(t, ctx, org)

		mod1 := svc.createModule(t, ctx, org)
		_, err := svc.Connections.Connect(ctx, connections.ConnectOptions{
			ConnectionType: connections.ModuleConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     mod1.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)

		hook := <-svc.WebhookEvents
		require.Equal(t, github.WebhookCreated, hook.Action)

		mod2 := svc.createModule(t, ctx, org)
		_, err = svc.Connections.Connect(ctx, connections.ConnectOptions{
			ConnectionType: connections.ModuleConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     mod2.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)

		hook = <-svc.WebhookEvents
		require.Equal(t, github.WebhookUpdated, hook.Action)

		ws1 := svc.createWorkspace(t, ctx, org)
		_, err = svc.Connections.Connect(ctx, connections.ConnectOptions{
			ConnectionType: connections.WorkspaceConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     ws1.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)

		hook = <-svc.WebhookEvents
		require.Equal(t, github.WebhookUpdated, hook.Action)

		ws2 := svc.createWorkspace(t, ctx, org)
		_, err = svc.Connections.Connect(ctx, connections.ConnectOptions{
			ConnectionType: connections.WorkspaceConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     ws2.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)

		hook = <-svc.WebhookEvents
		require.Equal(t, github.WebhookUpdated, hook.Action)

		t.Run("delete multiple connections", func(t *testing.T) {
			err = svc.Connections.Disconnect(ctx, connections.DisconnectOptions{
				ConnectionType: connections.WorkspaceConnection,
				ResourceID:     ws2.ID,
			})
			require.NoError(t, err)

			err = svc.Connections.Disconnect(ctx, connections.DisconnectOptions{
				ConnectionType: connections.WorkspaceConnection,
				ResourceID:     ws1.ID,
			})
			require.NoError(t, err)

			err := svc.Connections.Disconnect(ctx, connections.DisconnectOptions{
				ConnectionType: connections.ModuleConnection,
				ResourceID:     mod2.ID,
			})
			require.NoError(t, err)

			err = svc.Connections.Disconnect(ctx, connections.DisconnectOptions{
				ConnectionType: connections.ModuleConnection,
				ResourceID:     mod1.ID,
			})
			require.NoError(t, err)

			// webhook should now have been deleted from github
			hook = <-svc.WebhookEvents
			require.Equal(t, github.WebhookDeleted, hook.Action)

		})
	})
}
