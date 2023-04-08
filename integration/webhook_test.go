package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/github"
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
	t.Parallel()

	repo := cloud.NewTestRepo()

	// create an otf daemon with a fake github backend, ready to sign in a user,
	// serve up a repo and its contents via tarball. And register a callback to
	// test receipt of commit statuses
	svc := setup(t, nil,
		github.WithRepo(repo),
		github.WithRefs("tags/v0.0.1", "tags/v0.0.2", "tags/v0.1.0"),
		github.WithArchive(readFile(t, "./testdata/github.tar.gz")),
	)
	_, ctx := svc.createUserCtx(t, ctx)
	org := svc.createOrganization(t, ctx)

	// create and connect first workspace
	browser := createBrowserCtx(t)
	err := chromedp.Run(browser, chromedp.Tasks{
		createGithubVCSProviderTasks(t, svc.Hostname(), org.Name, "github"),

		createWorkspace(t, svc.Hostname(), org.Name, "workspace-1"),
		connectWorkspaceTasks(t, svc.Hostname(), org.Name, "workspace-1"),
	})
	require.NoError(t, err)

	// webhook should now have been registered with github
	require.True(t, svc.HasWebhook())

	// create and connect second workspace
	err = chromedp.Run(ctx, chromedp.Tasks{
		createWorkspace(t, svc.Hostname(), org.Name, "workspace-2"),
		connectWorkspaceTasks(t, svc.Hostname(), org.Name, "workspace-2"),
	})
	require.NoError(t, err)

	// second workspace re-uses same webhook on github
	require.True(t, svc.HasWebhook())

	// disconnect second workspace
	err = chromedp.Run(ctx, disconnectWorkspaceTasks(t, svc.Hostname(), org.Name, "workspace-2"))
	require.NoError(t, err)

	// first workspace is still connected, so webhook should still be configured
	// on github
	require.True(t, svc.HasWebhook())

	// disconnect first workspace
	err = chromedp.Run(ctx, disconnectWorkspaceTasks(t, svc.Hostname(), org.Name, "workspace-1"))
	require.NoError(t, err)

	// No more workspaces are connected to repo, so webhook should have been
	// deleted
	require.False(t, svc.HasWebhook())
}
