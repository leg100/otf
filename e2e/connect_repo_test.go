package e2e

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	gogithub "github.com/google/go-github/v41/github"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

// TestConnectRepo tests VCS integration, creating a VCS provider and connecting
// a workspace to a VCS repo.
func TestConnectRepo(t *testing.T) {
	addBuildsToPath(t)

	org := uuid.NewString()
	user := cloud.User{
		Name: "connect-repo-user",
		Teams: []cloud.Team{
			{
				Name:         "owners",
				Organization: org,
			},
		},
		Organizations: []string{org},
	}

	repo := otf.NewTestRepo()
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
	url := "https://" + hostname
	workspaceName := "workspace-connect"

	// create browser
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	// create timeout
	ctx, cancel = context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err = chromedp.Run(ctx, chromedp.Tasks{
		githubLoginTasks(t, hostname, user.Name),
		createGithubVCSProviderTasks(t, url, org, "github"),
		// create workspace via UI
		createWorkspaceTasks(t, hostname, org, workspaceName),
		// connect workspace to vcs repo
		chromedp.Tasks{
			// go to workspace
			chromedp.Navigate(path.Join(url, "organizations", org, "workspaces", workspaceName)),
			screenshot(t),
			// navigate to workspace settings
			chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
			screenshot(t),
			// click connect button
			chromedp.Click(`//button[text()='Connect to VCS']`, chromedp.NodeVisible),
			screenshot(t),
			// select provider
			chromedp.Click(`//a[normalize-space(text())='github']`, chromedp.NodeVisible),
			screenshot(t),
			// connect to first repo in list (there should only be one)
			chromedp.Click(`//div[@class='content-list']//button[text()='connect']`, chromedp.NodeVisible),
			screenshot(t),
			// confirm connected
			matchText(t, ".flash-success", "connected workspace to repo"),
		},
		// we can now start a run via the web ui, which'll retrieve the tarball from
		// the fake github server
		startRunTasks(t, hostname, org, workspaceName),
	})
	require.NoError(t, err)

	// Now we test the webhook functionality by sending an event to the daemon
	// (which would usually be triggered by a git push to github). The event
	// should trigger a run on the workspace.

	// otfd should have registered a webhook with the github server
	require.NotNil(t, daemon.githubServer.WebhookURL)
	require.NotNil(t, daemon.githubServer.WebhookSecret)

	// generate push event using template
	pushTpl, err := os.ReadFile("fixtures/github_push.json")
	require.NoError(t, err)
	push := fmt.Sprintf(string(pushTpl), repo.Identifier)

	// send push event
	sendGithubPushEvent(t, []byte(push), *daemon.githubServer.WebhookURL, *daemon.githubServer.WebhookSecret)

	// commit-triggered run should appear as latest run on workspace
	err = chromedp.Run(ctx, chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(fmt.Sprintf("%s/organizations/%s/workspaces/%s", url, org, workspaceName)),
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
		require.Equal(t, "no changes", *status.Description)
	}

	// Clean up after ourselves by deleting the vcs provider
	okDialog(t, ctx)
	err = chromedp.Run(ctx, chromedp.Tasks{
		// go to org
		chromedp.Navigate(path.Join(url, "organizations", org)),
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
