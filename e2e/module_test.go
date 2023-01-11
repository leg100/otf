package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/chromedp/chromedp"
	gogithub "github.com/google/go-github/v41/github"
	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

// TestModule tests publishing a module, first via the UI and then via a webhook
// event, and then invokes a terraform run that sources the module.
func TestModule(t *testing.T) {
	addBuildsToPath(t)

	name := "mod"
	provider := "aws"
	repo := cloud.NewTestModuleRepo(provider, name)

	org := uuid.NewString()
	user := cloud.User{
		Name: uuid.NewString(),
		Teams: []cloud.Team{
			{
				Name:         "owners",
				Organization: org,
			},
		},
		Organizations: []string{org},
	}

	// create an otf daemon with a fake github backend, ready to sign in a user,
	// serve up a repo and its contents via tarball. And register a callback to
	// test receipt of commit statuses
	daemon := &daemon{}
	daemon.withGithubUser(&user)
	daemon.withGithubRepo(repo)
	daemon.withGithubRefs("tags/v0.0.1", "tags/v0.0.2", "tags/v0.1.0")

	// create a tarball containing the module and seed our fake github with it
	tarball, err := os.ReadFile("./fixtures/github.module.tar.gz")
	require.NoError(t, err)
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
		githubLoginTasks(t, hostname, user.Name),
		createGithubVCSProviderTasks(t, url, org, "github"),
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

	// v1.0.0 should appear as latest module on workspace
	err = chromedp.Run(ctx, chromedp.Tasks{
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
	err = chromedp.Run(ctx, createWorkspaceTasks(t, hostname, org, workspaceName))
	require.NoError(t, err)

	// generate some terraform config that sources our module
	root := newRootModule(t, hostname, org, workspaceName)
	config := fmt.Sprintf(`
module "mod" {
  source  = "%s/%s/%s/%s"
  version = "1.0.0"
}
`, hostname, org, name, provider)
	err = os.WriteFile(filepath.Join(root, "sourcing.tf"), []byte(config), 0o600)
	require.NoError(t, err)

	// run terraform locally
	err = chromedp.Run(ctx, terraformLoginTasks(t, hostname))
	require.NoError(t, err)

	cmd := exec.Command("terraform", "init", "-no-color")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)

	cmd = exec.Command("terraform", "plan", "-no-color")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	require.Contains(t, string(out), "Plan: 2 to add, 0 to change, 0 to destroy.")

	cmd = exec.Command("terraform", "apply", "-no-color", "-auto-approve")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	require.Contains(t, string(out), "Apply complete! Resources: 2 added, 0 changed, 0 destroyed.")
}
