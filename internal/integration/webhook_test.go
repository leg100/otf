package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/require"
)

// TestWebhook tests webhook functionality. Two workspaces are created and
// connected to a repository. This should result in the creation of a webhook
// on github. Both workspaces are then disconnected. This should result in the
// deletion of the webhook from github.
//
// Functionality specific to VCS events triggering webhooks and in turn
// triggering workspace runs and publishing module versions in tested in other
// E2E tests.
func TestWebhook(t *testing.T) {
	integrationTest(t)

	repo := cloud.NewTestRepo()

	// create an otf daemon with a fake github backend, ready to sign in a user,
	// serve up a repo and its contents via tarball. And register a callback to
	// test receipt of commit statuses
	daemon, org, ctx := setup(t, nil,
		github.WithRepo(repo),
		github.WithRefs("tags/v0.0.1", "tags/v0.0.2", "tags/v0.1.0"),
		github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
	)

	// create and connect first workspace
	browser.Run(t, ctx, chromedp.Tasks{
		createGithubVCSProviderTasks(t, daemon.Hostname(), org.Name, "github"),
		createWorkspace(t, daemon.Hostname(), org.Name, "workspace-1"),
		connectWorkspaceTasks(t, daemon.Hostname(), org.Name, "workspace-1"),
	})

	// webhook should now have been registered with github
	require.True(t, daemon.HasWebhook())

	// create and connect second workspace
	browser.Run(t, ctx, chromedp.Tasks{
		createWorkspace(t, daemon.Hostname(), org.Name, "workspace-2"),
		connectWorkspaceTasks(t, daemon.Hostname(), org.Name, "workspace-2"),
	})

	// second workspace re-uses same webhook on github
	require.True(t, daemon.HasWebhook())

	// disconnect second workspace
	browser.Run(t, ctx, disconnectWorkspaceTasks(t, daemon.Hostname(), org.Name, "workspace-2"))

	// first workspace is still connected, so webhook should still be configured
	// on github
	require.True(t, daemon.HasWebhook())

	// disconnect first workspace
	browser.Run(t, ctx, disconnectWorkspaceTasks(t, daemon.Hostname(), org.Name, "workspace-1"))

	// No more workspaces are connected to repo, so webhook should have been
	// deleted
	require.False(t, daemon.HasWebhook())
}
