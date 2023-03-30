package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	gogithub "github.com/google/go-github/v41/github"
	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

// TestConnectRepo tests connecting a workspace to a VCS repository, pushing a
// git commit which should trigger a run on the workspace.
func TestConnectRepo(t *testing.T) {
	org, workspace := setup(t)

	user := cloud.User{
		Name: uuid.NewString(),
		Teams: []cloud.Team{
			{
				Name:         "owners",
				Organization: org,
			},
		},
	}

	repo := cloud.NewTestRepo()
	tarball, err := os.ReadFile("../testdata/github.tar.gz")
	require.NoError(t, err)

	// create an otf daemon with a fake github backend, ready to sign in a user,
	// serve up a repo and its contents via tarball. And register a callback to
	// test receipt of commit statuses
	daemon := &daemon{}
	daemon.withGithubUser(&user)
	daemon.withGithubRepo(repo)
	daemon.withGithubTarball(tarball)

	statuses := make(chan *gogithub.StatusEvent, 10)
	daemon.registerStatusCallback(func(status *gogithub.StatusEvent) {
		statuses <- status
	})

	hostname := daemon.start(t)

	// create browser
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	// create timeout
	ctx, cancel = context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err = chromedp.Run(ctx, chromedp.Tasks{
		githubLoginTasks(t, hostname, user.Name),
		createGithubVCSProviderTasks(t, hostname, org, "github"),
		createWorkspaceTasks(t, hostname, org, workspace),
		connectWorkspaceTasks(t, hostname, org, workspace),
		// we can now start a run via the web ui, which'll retrieve the tarball from
		// the fake github server
		startRunTasks(t, hostname, org, workspace),
	})
	require.NoError(t, err)

	// Now we test the webhook functionality by sending an event to the daemon
	// (which would usually be triggered by a git push to github). The event
	// should trigger a run on the workspace.

	// otfd should have registered a webhook with the github server
	require.True(t, daemon.githubServer.HasWebhook())

	// generate push event using template
	pushTpl, err := os.ReadFile("fixtures/github_push.json")
	require.NoError(t, err)
	push := fmt.Sprintf(string(pushTpl), repo)

	// send push event
	sendGithubPushEvent(t, []byte(push), *daemon.githubServer.HookEndpoint, *daemon.githubServer.HookSecret)

	// commit-triggered run should appear as latest run on workspace
	err = chromedp.Run(ctx, chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(workspacePath(hostname, org, workspace)),
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
	okDialog(t, ctx)
	err = chromedp.Run(ctx, chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(workspacePath(hostname, org, workspace)),
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
		matchText(t, ".flash-success", "deleted workspace: "+workspace),
		//
		// delete vcs provider
		//
		// go to org
		chromedp.Navigate(organizationPath(hostname, org)),
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
