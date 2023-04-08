package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/chromedp/chromedp"
	gogithub "github.com/google/go-github/v41/github"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/github"
	"github.com/stretchr/testify/require"
)

// TestModuleE2E tests publishing a module, first via the UI and then via a webhook
// event, and then invokes a terraform run that sources the module.
func TestModuleE2E(t *testing.T) {
	t.Parallel()

	// create an otf daemon with a fake github backend, ready to serve up a repo
	// and its contents via tarball. And register a callback to test receipt of
	// commit statuses
	statuses := make(chan *gogithub.StatusEvent, 10)
	repo := cloud.NewTestModuleRepo("aws", "mod")
	svc := setup(t, nil,
		github.WithRepo(repo),
		github.WithRefs("tags/v0.0.1", "tags/v0.0.2", "tags/v0.1.0"),
		github.WithArchive(readFile(t, "./fixtures/github.module.tar.gz")),
		github.WithStatusCallback(func(status *gogithub.StatusEvent) {
			statuses <- status
		}),
	)
	user, ctx := svc.createUserCtx(t, ctx)
	org := svc.createOrganization(t, ctx)

	var moduleURL string // captures url for module page
	browser := createBrowserCtx(t)
	err := chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ctx, svc.Hostname(), user.Username, svc.Secret),
		createGithubVCSProviderTasks(t, svc.Hostname(), org.Name, "github"),
		// publish module
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(organizationURL(svc.Hostname(), org.Name)),
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
			matchText(t, ".flash-success", "published module: mod"),
			// TODO: confirm versions are populated
			// capture module url so we can visit it later
			chromedp.Location(&moduleURL),
		},
	})
	require.NoError(t, err)

	// Now we test the webhook functionality by sending an event to the daemon
	// (which would usually be triggered by a git push to github). The event
	// should trigger a module version to be published.

	// otfd should have registered a webhook with the github server
	require.True(t, svc.HasWebhook())

	// generate and send push tag event for v1.0.0
	pushTpl := readFile(t, "fixtures/github_push_tag.json")
	push := fmt.Sprintf(string(pushTpl), "v1.0.0", repo)
	sendGithubPushEvent(t, []byte(push), *svc.HookEndpoint, *svc.HookSecret)

	// v1.0.0 should appear as latest module on workspace
	err = chromedp.Run(browser, chromedp.Tasks{
		// go to module
		chromedp.Navigate(moduleURL),
		screenshot(t),
		reloadUntilVisible(`//select[@id="version"]/option[@selected]`),
		screenshot(t),
	})
	require.NoError(t, err)

	// Now run terraform with some config that sources the module. First we need
	// a workspace...
	workspaceName := "module-test"
	err = chromedp.Run(browser, createWorkspace(t, svc.Hostname(), org.Name, workspaceName))
	require.NoError(t, err)

	// generate some terraform config that sources our module
	root := newRootModule(t, svc.Hostname(), org.Name, workspaceName)
	config := fmt.Sprintf(`
module "mod" {
  source  = "%s/%s/%s/%s"
  version = "1.0.0"
}
`, svc.Hostname(), org.Name, "mod", "aws")
	err = os.WriteFile(filepath.Join(root, "sourcing.tf"), []byte(config), 0o600)
	require.NoError(t, err)

	// run terraform init, plan, and apply
	svc.tfcli(t, ctx, "init", root)
	out := svc.tfcli(t, ctx, "plan", root)
	require.Contains(t, out, "Plan: 2 to add, 0 to change, 0 to destroy.")
	out = svc.tfcli(t, ctx, "apply", root, "-auto-approve")
	require.Contains(t, string(out), "Apply complete! Resources: 2 added, 0 changed, 0 destroyed.")
}
