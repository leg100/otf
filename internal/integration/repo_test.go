package integration

import (
	"testing"

	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/require"
)

func TestRepo(t *testing.T) {
	integrationTest(t)

	t.Run("create multiple connections", func(t *testing.T) {
		daemon, org, ctx := setup(t, withGithubOption(github.WithRepo(vcs.NewMustRepo("test", "dummy"))))

		vcsprov := daemon.createVCSProvider(t, ctx, org, nil)

		mod1 := daemon.createModule(t, ctx, org)
		_, err := daemon.Connections.Connect(ctx, connections.ConnectOptions{
			VCSProviderID: vcsprov.ID,
			ResourceID:    mod1.ID,
			RepoPath:      vcs.NewMustRepo("test", "dummy"),
		})
		require.NoError(t, err)

		hook := <-daemon.WebhookEvents
		require.Equal(t, github.WebhookCreated, hook.Action)

		mod2 := daemon.createModule(t, ctx, org)
		_, err = daemon.Connections.Connect(ctx, connections.ConnectOptions{
			VCSProviderID: vcsprov.ID,
			ResourceID:    mod2.ID,
			RepoPath:      vcs.NewMustRepo("test", "dummy"),
		})
		require.NoError(t, err)

		hook = <-daemon.WebhookEvents
		require.Equal(t, github.WebhookUpdated, hook.Action)

		ws1 := daemon.createWorkspace(t, ctx, org)
		_, err = daemon.Connections.Connect(ctx, connections.ConnectOptions{
			VCSProviderID: vcsprov.ID,
			ResourceID:    ws1.ID,
			RepoPath:      vcs.NewMustRepo("test", "dummy"),
		})
		require.NoError(t, err)

		hook = <-daemon.WebhookEvents
		require.Equal(t, github.WebhookUpdated, hook.Action)

		ws2 := daemon.createWorkspace(t, ctx, org)
		_, err = daemon.Connections.Connect(ctx, connections.ConnectOptions{
			VCSProviderID: vcsprov.ID,
			ResourceID:    ws2.ID,
			RepoPath:      vcs.NewMustRepo("test", "dummy"),
		})
		require.NoError(t, err)

		hook = <-daemon.WebhookEvents
		require.Equal(t, github.WebhookUpdated, hook.Action)

		t.Run("delete multiple connections", func(t *testing.T) {
			err = daemon.Connections.Disconnect(ctx, connections.DisconnectOptions{
				ResourceID: ws2.ID,
			})
			require.NoError(t, err)

			err = daemon.Connections.Disconnect(ctx, connections.DisconnectOptions{
				ResourceID: ws1.ID,
			})
			require.NoError(t, err)

			err := daemon.Connections.Disconnect(ctx, connections.DisconnectOptions{
				ResourceID: mod2.ID,
			})
			require.NoError(t, err)

			err = daemon.Connections.Disconnect(ctx, connections.DisconnectOptions{
				ResourceID: mod1.ID,
			})
			require.NoError(t, err)

			// webhook should now have been deleted from github
			hook = <-daemon.WebhookEvents
			require.Equal(t, github.WebhookDeleted, hook.Action)
		})
	})
}
