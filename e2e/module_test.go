package e2e

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/chromedp/chromedp"
	gogithub "github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

// TestModule tests publishing a module.
func TestModule(t *testing.T) {
	addBuildsToPath(t)

	name := "mod"
	provider := "aws"
	repo := otf.NewTestModuleRepo(provider, name)

	user := otf.NewTestUser(t)
	org := user.Username() // we'll be using user's personal organization
	tarball, err := os.ReadFile("../testdata/github.tar.gz")
	require.NoError(t, err)

	// create an otf daemon with a fake github backend, ready to sign in a user,
	// serve up a repo and its contents via tarball. And register a callback to
	// test receipt of commit statuses
	daemon := &daemon{}
	daemon.withGithubUser(user)
	daemon.withGithubRepo(repo)
	daemon.withGithubRefs("tags/v0.0.1", "tags/v0.0.2", "tags/v0.1.0")
	daemon.withGithubTarball(tarball)

	statuses := make(chan *gogithub.StatusEvent, 10)
	daemon.registerStatusCallback(func(status *gogithub.StatusEvent) {
		statuses <- status
	})

	hostname := daemon.start(t)
	url := "https://" + hostname

	// create browser
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	var moduleURL string // captures url for module page
	err = chromedp.Run(ctx, chromedp.Tasks{
		githubLoginTasks(t, hostname, user.Username()),
		createGithubVCSProviderTasks(t, url, org),
		// publish module
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(path.Join(url, "organizations", org)),
			screenshot(t),
			// go to modules
			chromedp.Click("#modules > a", chromedp.NodeVisible),
			screenshot(t),
			// click publish button
			chromedp.Click(`//button[text()='Publish']`, chromedp.NodeVisible),
			screenshot(t),
			// select provider
			chromedp.Click(`//button[text()='connect']`, chromedp.NodeVisible),
			screenshot(t),
			// connect to first repo in list (there should only be one)
			chromedp.Click(`//div[@class='content-list']//button[text()='connect']`, chromedp.NodeVisible),
			screenshot(t),
			// confirm module details
			chromedp.Click(`//button[text()='connect']`, chromedp.NodeVisible),
			// flash message indicates success
			matchText(t, ".flash-success", "published module: "+name),
			// TODO: confirm versions are populated
			// capture module url so we can visit it later
			chromedp.Location(&moduleURL),
		},
	})
	require.NoError(t, err)

	// Now we test the webhook functionality by sending an event to the daemon
	// (which would usually be triggered by a git push to github). The event
	// should trigger a run on the workspace.

	// otfd should have registered a webhook with the github server
	require.NotNil(t, daemon.githubServer.WebhookURL)
	require.NotNil(t, daemon.githubServer.WebhookSecret)

	// generate and send push tag event for v1.0.0
	pushTpl, err := os.ReadFile("fixtures/github_push_tag.json")
	require.NoError(t, err)
	push := fmt.Sprintf(string(pushTpl), "v1.0.0", repo.Identifier)
	sendGithubPushEvent(t, []byte(push), *daemon.githubServer.WebhookURL, *daemon.githubServer.WebhookSecret)

	// TODO: need to poll for new version, reloading the browser.

	// v1.0.0 should appear as latest module on workspace
	err = chromedp.Run(ctx, chromedp.Tasks{
		// go to module
		chromedp.Navigate(moduleURL),
		screenshot(t),
		reloadUntilVisible(`//select[@id="version"]/option[@selected]`),
		screenshot(t),
	})
	require.NoError(t, err)
}
