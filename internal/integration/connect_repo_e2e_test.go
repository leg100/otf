package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/require"
)

// TestConnectRepoE2E demonstrates connecting a workspace to a VCS repository, pushing a
// git commit which triggers a run on the workspace.
func TestConnectRepoE2E(t *testing.T) {
	integrationTest(t)

	// create an otf daemon with a fake github backend, serve up a repo and its
	// contents via tarball. And register a callback to test receipt of commit
	// statuses
	repo := cloud.NewTestRepo()
	daemon, org, ctx := setup(t, nil,
		github.WithRepo(repo),
		github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
	)

	browser.Run(t, ctx, chromedp.Tasks{
		createGithubVCSProviderTasks(t, daemon.Hostname(), org.Name, "github"),
		createWorkspace(t, daemon.Hostname(), org.Name, "my-test-workspace"),
		connectWorkspaceTasks(t, daemon.Hostname(), org.Name, "my-test-workspace"),
		// we can now start a run via the web ui, which'll retrieve the tarball from
		// the fake github server
		startRunTasks(t, daemon.Hostname(), org.Name, "my-test-workspace", run.PlanAndApplyOperation),
	})

	// Now we test the webhook functionality by sending an event to the daemon
	// (which would usually be triggered by a git push to github). The event
	// should trigger a run on the workspace.

	// generate and send push event
	push := testutils.ReadFile(t, "fixtures/github_push.json")
	daemon.SendEvent(t, github.PushEvent, push)

	// commit-triggered run should appear as latest run on workspace
	browser.Run(t, ctx, chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(workspaceURL(daemon.Hostname(), org.Name, "my-test-workspace")),
		screenshot(t),
		// commit should match that of push event
		chromedp.WaitVisible(`//div[@id='latest-run']//span[@class='commit' and text()='#42d6fc7']`),
		screenshot(t),
	})

	// github should receive three pending status updates followed by a final
	// update with details of planned resources
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	got := daemon.GetStatus(t, ctx)
	require.Equal(t, "success", got.GetState())
	require.Equal(t, "planned: +0/~0/−0", got.GetDescription())

	// Clean up after ourselves by disconnecting the workspace and deleting the
	// workspace and vcs provider
	browser.Run(t, ctx, chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(workspaceURL(daemon.Hostname(), org.Name, "my-test-workspace")),
		screenshot(t),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`),
		screenshot(t),
		// click disconnect button
		chromedp.Click(`//button[@id='disconnect-workspace-repo-button']`),
		screenshot(t),
		// confirm disconnected
		matchText(t, ".flash-success", "disconnected workspace from repo", chromedp.ByQuery),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`),
		screenshot(t),
		// delete workspace
		chromedp.Click(`//button[@id='delete-workspace-button']`),
		screenshot(t),
		// confirm deletion
		matchText(t, ".flash-success", "deleted workspace: my-test-workspace", chromedp.ByQuery),
		//
		// delete vcs provider
		//
		// go to org
		chromedp.Navigate(organizationURL(daemon.Hostname(), org.Name)),
		screenshot(t),
		// go to vcs providers
		chromedp.Click("#vcs_providers > a"),
		screenshot(t),
		// click delete button for one and only vcs provider
		chromedp.Click(`//button[text()='delete']`),
		screenshot(t),
		matchText(t, ".flash-success", "deleted provider: github", chromedp.ByQuery),
	})
}
