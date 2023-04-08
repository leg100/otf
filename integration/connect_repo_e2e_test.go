package integration

import (
	"fmt"
	"testing"

	"github.com/chromedp/chromedp"
	gogithub "github.com/google/go-github/v41/github"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/github"
	"github.com/stretchr/testify/require"
)

// TestConnectRepoE2E tests connecting a workspace to a VCS repository, pushing a
// git commit which should trigger a run on the workspace.
func TestConnectRepoE2E(t *testing.T) {
	t.Parallel()

	// create an otf daemon with a fake github backend, ready to sign in a user,
	// serve up a repo and its contents via tarball. And register a callback to
	// test receipt of commit statuses
	repo := cloud.NewTestRepo()
	statuses := make(chan *gogithub.StatusEvent, 10)
	daemon := setup(t, nil,
		github.WithRepo(repo),
		github.WithArchive(readFile(t, "../testdata/github.tar.gz")),
		github.WithStatusCallback(func(status *gogithub.StatusEvent) {
			statuses <- status
		}),
	)
	user, ctx := daemon.createUserCtx(t, ctx)
	org := daemon.createOrganization(t, ctx)

	browser := createBrowserCtx(t)
	err := chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ctx, daemon.Hostname(), user.Username, daemon.Secret),
		createGithubVCSProviderTasks(t, daemon.Hostname(), org.Name, "github"),
		createWorkspace(t, daemon.Hostname(), org.Name, "my-test-workspace"),
		connectWorkspaceTasks(t, daemon.Hostname(), org.Name, "my-test-workspace"),
		// we can now start a run via the web ui, which'll retrieve the tarball from
		// the fake github server
		startRunTasks(t, daemon.Hostname(), org.Name, "my-test-workspace"),
	})
	require.NoError(t, err)

	// Now we test the webhook functionality by sending an event to the daemon
	// (which would usually be triggered by a git push to github). The event
	// should trigger a run on the workspace.

	// otfd should have registered a webhook with the github server
	require.True(t, daemon.HasWebhook())

	// generate push event using template
	pushTpl := readFile(t, "fixtures/github_push.json")
	push := fmt.Sprintf(string(pushTpl), repo)

	// send push event
	sendGithubPushEvent(t, []byte(push), *daemon.HookEndpoint, *daemon.HookSecret)

	// commit-triggered run should appear as latest run on workspace
	err = chromedp.Run(browser, chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(workspacePath(daemon.Hostname(), org.Name, "my-test-workspace")),
		screenshot(t),
		// commit should match that of push event
		chromedp.WaitVisible(`//div[@id='latest-run']//span[@class='commit' and text()='#42d6fc7']`),
		screenshot(t),
	})
	require.NoError(t, err)

	// check github received commit statuses
	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	case status := <-statuses:
		require.Equal(t, "pending", *status.State)
	}

	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	case status := <-statuses:
		require.Equal(t, "pending", *status.State)
	}

	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	case status := <-statuses:
		require.Equal(t, "pending", *status.State)
	}

	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	case status := <-statuses:
		require.Equal(t, "success", *status.State)
		require.Equal(t, "planned: +0/~0/−0", *status.Description)
	}

	// Clean up after ourselves by disconnecting the workspace and deleting the
	// workspace and vcs provider
	okDialog(t, browser)
	err = chromedp.Run(browser, chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(workspacePath(daemon.Hostname(), org.Name, "my-test-workspace")),
		screenshot(t),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
		screenshot(t),
		// click disconnect button
		chromedp.Click(`//button[@id='disconnect-workspace-repo-button']`, chromedp.NodeVisible),
		screenshot(t),
		// confirm disconnected
		matchText(t, ".flash-success", "disconnected workspace from repo"),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
		screenshot(t),
		// delete workspace
		chromedp.Click(`//button[text()='Delete workspace']`, chromedp.NodeVisible),
		screenshot(t),
		// confirm deletion
		matchText(t, ".flash-success", "deleted workspace: my-test-workspace"),
		//
		// delete vcs provider
		//
		// go to org
		chromedp.Navigate(organizationPath(daemon.Hostname(), org.Name)),
		screenshot(t),
		// go to vcs providers
		chromedp.Click("#vcs_providers > a", chromedp.NodeVisible),
		screenshot(t),
		// click delete button for one and only vcs provider
		chromedp.Click(`//button[text()='delete']`, chromedp.NodeVisible),
		screenshot(t),
		matchText(t, ".flash-success", "deleted provider: github"),
	})
	require.NoError(t, err)
}
