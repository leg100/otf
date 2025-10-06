package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/vcs"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestModuleE2E tests publishing a module, first via the UI and then via a webhook
// event, and then invokes a terraform run that sources the module.
func TestModuleE2E(t *testing.T) {
	integrationTest(t)

	// create an otf daemon with a fake github backend, ready to serve up a repo
	// and its contents via tarball.
	repo := vcs.NewRandomModuleRepo("aws", "mod")
	svc, org, ctx := setup(t, withGithubOptions(
		github.WithRepo(repo),
		github.WithRefs("tags/v0.0.1", "tags/v0.0.2", "tags/v0.1.0"),
		github.WithArchive(testutils.ReadFile(t, "./fixtures/github.module.tar.gz")),
	))
	// create vcs provider for module to authenticate to github backend
	provider := svc.createVCSProvider(t, ctx, org, nil)

	var moduleURL string // captures url for module page

	browser.New(t, ctx, func(page playwright.Page) {
		// publish module
		// go to org
		_, err := page.Goto(organizationURL(svc.System.Hostname(), org.Name))
		require.NoError(t, err)

		// go to modules
		err = page.Locator("#menu-item-modules > a").Click()
		require.NoError(t, err)
		screenshot(t, page, "modules_list")

		// click publish button
		err = page.Locator(`//button[text()='Publish']`).Click()
		require.NoError(t, err)
		screenshot(t, page, "modules_select_provider")

		// select provider
		err = page.Locator(`//button[text()='Select']`).Click()
		require.NoError(t, err)
		screenshot(t, page, "modules_select_repo")

		// connect to first repo in list (there should only be one)
		err = page.Locator(`//tr[@id='item-repo-` + repo.String() + `']//button[text()='Connect']`).Click()
		require.NoError(t, err)

		// flash message indicates success
		err = expect.Locator(page.GetByRole("alert")).ToHaveText(`published module: mod`)
		require.NoError(t, err)
		screenshot(t, page, "newly_created_module_page")

		// capture module url so we can visit it later
		moduleURL = page.URL()

		// confirm versions are populated
		err = expect.Locator(page.Locator(`//select[@id='version']/option[text()='0.0.1']`)).ToBeEnabled()
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//select[@id='version']/option[text()='0.0.2']`)).ToBeEnabled()
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//select[@id='version']/option[text()='0.1.0']`)).ToBeEnabled()
		require.NoError(t, err)

		// should show vcs repo source
		err = expect.Locator(page.Locator(`//span[@id='vcs-repo']`)).ToHaveText(regexp.MustCompile(`.*/terraform-aws-mod`))
		require.NoError(t, err)

		// should show usage
		err = expect.Locator(page.Locator(`//div[@id='usage']`)).ToHaveText(fmt.Sprintf(`
module "mod" {
	source = "%s/%s/mod/aws"
	version = "0.1.0"
}
`, svc.System.Hostname(), org.Name))
		require.NoError(t, err)
	})

	// Now we test the webhook functionality by sending an event to the daemon
	// (which would usually be triggered by a git push to github). The event
	// should trigger a module version to be published.

	// generate and send push tag event for v1.0.0
	pushTpl := testutils.ReadFile(t, "fixtures/github_push_tag.json")
	push := fmt.Sprintf(string(pushTpl), "v1.0.0", repo.Name(), repo.Owner())
	svc.SendEvent(t, github.PushEvent, []byte(push))

	workspaceName := "module-test"
	browser.New(t, ctx, func(page playwright.Page) {
		// v1.0.0 should appear as latest module on workspace
		// go to module
		_, err := page.Goto(moduleURL)
		require.NoError(t, err)

		reloadUntilEnabled(t, page, `//select[@id="version"]/option[@selected]`)

		// Now run terraform with some config that sources the module. First we need
		// a workspace...
		createWorkspace(t, page, svc.System.Hostname(), org.Name, workspaceName)
	})

	// generate some terraform config that sources our module
	root := newRootModule(t, svc.System.Hostname(), org.Name, workspaceName)
	config := fmt.Sprintf(`
module "mod" {
  source  = "%s/%s/%s/%s"
  version = "1.0.0"
}
`, svc.System.Hostname(), org.Name, "mod", "aws")
	err := os.WriteFile(filepath.Join(root, "sourcing.tf"), []byte(config), 0o600)
	require.NoError(t, err)

	// run terraform init, plan, and apply
	svc.engineCLI(t, ctx, "", "init", root)
	out := svc.engineCLI(t, ctx, "", "plan", root)
	require.Contains(t, out, "Plan: 2 to add, 0 to change, 0 to destroy.")
	out = svc.engineCLI(t, ctx, "", "apply", root, "-auto-approve")
	require.Contains(t, string(out), "Apply complete! Resources: 2 added, 0 changed, 0 destroyed.")

	// delete vcs provider and visit the module page; it should be no longer
	// connected. Then delete the module.
	_, err = svc.VCSProviders.Delete(ctx, provider.ID)
	require.NoError(t, err)

	browser.New(t, ctx, func(page playwright.Page) {
		// go to org
		_, err = page.Goto(organizationURL(svc.System.Hostname(), org.Name))
		require.NoError(t, err)

		// go to modules
		err = page.Locator("#menu-item-modules > a").Click()
		require.NoError(t, err)

		// select existing module
		err = page.Locator(`//tr[@id='mod-item-mod']/td[1]/a`).Click()
		require.NoError(t, err)

		// confirm no longer connected
		err = expect.Locator(page.Locator(`//span[@id='vcs-repo']`)).ToBeHidden()
		require.NoError(t, err)

		// delete module
		err = page.Locator(`//button[text()='Delete module']`).Click()
		require.NoError(t, err)

		// flash message indicates success
		err = expect.Locator(page.GetByRole("alert")).ToHaveText(`deleted module: mod`)
		require.NoError(t, err)
	})
}
