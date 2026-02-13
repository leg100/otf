package integration

import (
	"testing"

	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestWebhook tests webhook functionality. Two workspaces are created and
// connected to a repository. This should result in the creation of a webhook
// on github. Both workspaces are then disconnected. This should result in the
// deletion of the webhook from github.
//
// Functionality specific to VCS events triggering webhooks and in turn
// triggering workspace runs and publishing module versions is tested in other
// E2E tests.
func TestWebhook(t *testing.T) {
	integrationTest(t)

	repo := vcs.NewRandomRepo()

	// create otf daemon with fake github server, on which to create/delete
	// webhooks.
	daemon, org, ctx := setup(t, withGithubOption(github.WithRepo(repo)))
	// create vcs provider for authenticating to github backend
	provider := daemon.createVCSProvider(t, ctx, org, nil)

	// create and connect first workspace
	browser.New(t, ctx, func(page playwright.Page) {
		createWorkspace(t, page, daemon.System.Hostname(), org.Name, "workspace-1")
		connectWorkspaceTasks(t, page, daemon.System.Hostname(), org.Name, "workspace-1", provider.String())

		// webhook should be registered with github
		hook := <-daemon.WebhookEvents
		require.Equal(t, github.WebhookCreated, hook.Action)

		// create and connect second workspace
		createWorkspace(t, page, daemon.System.Hostname(), org.Name, "workspace-2")
		connectWorkspaceTasks(t, page, daemon.System.Hostname(), org.Name, "workspace-2", provider.String())

		// second workspace re-uses same webhook on github
		hook = <-daemon.WebhookEvents
		require.Equal(t, github.WebhookUpdated, hook.Action)

		// disconnect second workspace
		disconnectWorkspaceTasks(t, page, daemon.System.Hostname(), org.Name, "workspace-2")

		// first workspace is still connected, so webhook should still be configured
		// on github
		require.True(t, daemon.HasWebhook())

		// disconnect first workspace
		disconnectWorkspaceTasks(t, page, daemon.System.Hostname(), org.Name, "workspace-1")

		// No more workspaces are connected to repo, so webhook should have been
		// deleted
		hook = <-daemon.WebhookEvents
		require.Equal(t, github.WebhookDeleted, hook.Action)
	})
}

// TestWebhook_Purger tests specifically the purging of webhooks in response to
// various events.
func TestWebhook_Purger(t *testing.T) {
	integrationTest(t)

	repo := vcs.NewRandomRepo()

	// create an otf daemon with a fake github backend, ready to sign in a user,
	// serve up a repo and its contents via tarball. And register a callback to
	// test receipt of commit statuses
	daemon, _, ctx := setup(t, withGithubOption(github.WithRepo(repo)))

	tests := []struct {
		name  string
		event func(*testing.T, organization.Name, resource.TfeID, resource.TfeID)
	}{
		{
			name: "delete organization",
			event: func(t *testing.T, org organization.Name, _, vcsProviderID resource.TfeID) {
				err := daemon.Organizations.Delete(ctx, org)
				require.NoError(t, err)
			},
		},
		{
			name: "delete vcs provider",
			event: func(t *testing.T, _ organization.Name, _, vcsProviderID resource.TfeID) {
				_, err := daemon.VCSProviders.Delete(ctx, vcsProviderID)
				require.NoError(t, err)
			},
		},
		{
			name: "delete workspace",
			event: func(t *testing.T, _ organization.Name, workspaceID, _ resource.TfeID) {
				_, err := daemon.Workspaces.Delete(ctx, workspaceID)
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create org, vcs provider, and workspace, and connect the
			// workspace to create a webhook on github
			org := daemon.createOrganization(t, ctx)
			provider := daemon.createVCSProvider(t, ctx, org, nil)
			ws, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
				Name:         new("workspace-1"),
				Organization: &org.Name,
				ConnectOptions: &workspace.ConnectOptions{
					VCSProviderID: &provider.ID,
					RepoPath:      &repo,
				},
			})
			require.NoError(t, err)

			// webhook should have been registered with github
			hook := <-daemon.WebhookEvents
			require.Equal(t, github.WebhookCreated, hook.Action)

			tt.event(t, org.Name, ws.ID, provider.ID)

			// webhook should now have been deleted from  github
			hook = <-daemon.WebhookEvents
			require.Equal(t, github.WebhookDeleted, hook.Action)
		})
	}
}
