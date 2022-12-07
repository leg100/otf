package e2e

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	gogithub "github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConnectRepo tests VCS integration, creating a VCS provider and connecting
// a workspace to a VCS repo.
func TestConnectRepo(t *testing.T) {
	addBuildsToPath(t)

	user := otf.NewTestUser(t)
	repo := otf.NewTestRepo()
	// test using user's personal organization
	org := user.Username()

	tarball, err := os.ReadFile("../testdata/github.tar.gz")
	require.NoError(t, err)

	// create an otf daemon with a fake github backend, ready to sign in a user,
	// serve up a repo and its contents via tarball. And register a callback to
	// test receipt of commit statuses
	daemon := &daemon{}
	daemon.withGithubUser(user)
	daemon.withGithubRepo(repo)
	daemon.withGithubTarball(tarball)

	statuses := make(chan *gogithub.StatusEvent, 10)
	daemon.registerStatusCallback(func(status *gogithub.StatusEvent) {
		statuses <- status
	})

	hostname := daemon.start(t)
	url := "https://" + hostname

	// create chrome instance
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	// create timeout
	ctx, cancel = context.WithTimeout(ctx, time.Minute)
	defer cancel()

	orgSelector := fmt.Sprintf("#item-organization-%s a", org)

	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		// login
		chromedp.Click(".login-button-github", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		// select org
		chromedp.Click(orgSelector, chromedp.NodeVisible),
		// select vcs providers
		chromedp.Click("#vcs_providers > a", chromedp.NodeVisible),
		ss.screenshot(t),
		// click 'New Github VCS Provider' button
		chromedp.Click(`//button[text()='New Github VCS Provider']`, chromedp.NodeVisible),
		ss.screenshot(t),
		// enter fake github token and name
		chromedp.Focus("input#token", chromedp.NodeVisible),
		input.InsertText("fake-github-personal-token"),
		chromedp.Focus("input#name"),
		input.InsertText("github"),
		ss.screenshot(t),
		// submit form to create provider
		chromedp.Submit("input#token"),
	})
	require.NoError(t, err)

	// create workspace via UI before we connect to a repo
	workspace := createWebWorkspace(t, allocator, url, org)
	workspaceSelector := fmt.Sprintf("#item-workspace-%s a", workspace)

	// capture flash message confirming workspace has been connected
	var workspaceConnected string

	err = chromedp.Run(ctx, chromedp.Tasks{
		// go to list of organizations
		chromedp.Navigate(url + "/organizations"),
		// select my org
		chromedp.Click(orgSelector, chromedp.NodeVisible),
		// go to list of workspaces
		chromedp.Click("#menu-item-workspaces > a", chromedp.ByQuery),
		// select my workspace
		chromedp.Click(workspaceSelector, chromedp.NodeVisible),
		// navigate to workspace settings
		chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
		ss.screenshot(t),
		// click connect button
		chromedp.Click(`//button[text()='Connect to VCS']`, chromedp.NodeVisible),
		ss.screenshot(t),
		// select provider
		chromedp.Click(`//a[normalize-space(text())='github']`, chromedp.NodeVisible),
		ss.screenshot(t),
		// connect to first repo in list (there should only be one)
		chromedp.Click(`//div[@class='content-list']//button[text()='connect']`, chromedp.NodeVisible),
		ss.screenshot(t),
		// confirm connected
		//
		// did not work on two occasions
		chromedp.Text(".flash-success", &workspaceConnected, chromedp.NodeVisible),
	})
	require.NoError(t, err)
	assert.Equal(t, "connected workspace to repo", strings.TrimSpace(workspaceConnected))

	// we can now start a run via the web ui, which'll retrieve the tarball from
	// the fake github server
	err = chromedp.Run(ctx, startRunTasks(t, hostname, org, workspace))
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

	// generate signature for push event
	mac := hmac.New(sha256.New, []byte(*daemon.githubServer.WebhookSecret))
	mac.Write([]byte(push))
	sig := mac.Sum(nil)

	req, err := http.NewRequest("POST", *daemon.githubServer.WebhookURL, strings.NewReader(push))
	require.NoError(t, err)
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("X-GitHub-Event", "push")
	req.Header.Add("X-Hub-Signature-256", "sha256="+hex.EncodeToString(sig))

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	if !assert.Equal(t, http.StatusAccepted, res.StatusCode) {
		response, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		t.Fatal(string(response))
	}

	// commit-triggered run should appear as latest run on workspace
	err = chromedp.Run(ctx, chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(fmt.Sprintf("%s/organizations/%s/workspaces/%s", url, org, workspace)),
		ss.screenshot(t),
		// commit should match that of push event
		chromedp.WaitVisible(`//div[@id='latest-run']//span[@class='commit' and text()='#42d6fc7']`),
		ss.screenshot(t),
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
}
