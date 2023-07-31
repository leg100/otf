package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/require"
)

// TestGithubTriggerE2E demonstrates github events triggering runs.
func TestGithubTriggerE2E(t *testing.T) {
	integrationTest(t)

	// 1. create workspace, connect workspace and set trigger strategy
	// 2. send github event:
	// 3. check run is spawned and check its configuration (plan-only etc)
	//
	// table tests with combinations of:
	// a. trigger strategy
	//  i. always
	//  ii. file trigger patterns
	//  iii. tags
	// b. github events
	//  i. commit push
	//  ii. tag push
	//  iii. pull request open, update
	//

	// create an OTF daemon with a fake github backend, serve up a repo and its
	// contents via tarball, and setup a fake pull request with a list of files
	// it has changed.
	repo := cloud.NewTestRepo()
	daemon, org, ctx := setup(t, nil,
		github.WithRepo(repo),
		github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
		github.WithPullRequest("2", "foo.tf", "bar.tf"),
	)

	browser.Run(t, ctx, chromedp.Tasks{
		createGithubVCSProviderTasks(t, daemon.Hostname(), org.Name, "github"),
		createWorkspace(t, daemon.Hostname(), org.Name, "my-test-workspace"),
		connectWorkspaceTasks(t, daemon.Hostname(), org.Name, "my-test-workspace"),
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
		matchText(t, "//div[@role='alert']", "disconnected workspace from repo"),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`),
		screenshot(t),
		// delete workspace
		chromedp.Click(`//button[@id='delete-workspace-button']`),
		screenshot(t),
		// confirm deletion
		matchText(t, "//div[@role='alert']", "deleted workspace: my-test-workspace"),
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
		matchText(t, "//div[@role='alert']", "deleted provider: github"),
	})
}
