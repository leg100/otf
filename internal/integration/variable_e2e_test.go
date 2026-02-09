package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestVariableE2E tests adding, updating and deleting workspace variables via the
// UI, and tests that variables are made available to runs.
func TestVariableE2E(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t)

	// Create variable in browser
	browser.New(t, ctx, func(page playwright.Page) {
		createWorkspace(t, page, svc.System.Hostname(), org.Name, "my-test-workspace")

		// go to workspace
		_, err := page.Goto(workspaceURL(svc.System.Hostname(), org.Name, "my-test-workspace"))
		require.NoError(t, err)

		// go to variables
		err = page.Locator(`//li[@id='menu-item-variables']/a`).Click()
		require.NoError(t, err)

		// confirm 'variables' submenu button is active
		err = expect.Locator(page.Locator(`//*[@id="menu-item-variables"]/a`)).ToHaveClass(`menu-active`)
		require.NoError(t, err)

		// click add variable button
		err = page.Locator(`//button[text()='Add variable']`).Click()
		require.NoError(t, err)

		// enter key
		err = page.Locator("input#key").Fill("foo")
		require.NoError(t, err)

		// enter value
		err = page.Locator("textarea#value").Fill("bar")
		require.NoError(t, err)

		// select terraform variable category
		err = page.Locator("input#terraform").Click()
		require.NoError(t, err)

		// submit form
		err = page.Locator(`//button[@id='save-variable-button']`).Click()
		require.NoError(t, err)

		// confirm variable added
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("added variable: foo")
		require.NoError(t, err)
	})

	// write some terraform config that declares and outputs the variable
	root := newRootModule(t, svc.System.Hostname(), org.Name, "my-test-workspace")
	config := `
variable "foo" {
  default = "overwrite_this"
}

output "foo" {
  value = var.foo
}
`
	err := os.WriteFile(filepath.Join(root, "foo.tf"), []byte(config), 0o600)
	require.NoError(t, err)

	// run terraform init, plan, and apply
	svc.engineCLI(t, ctx, "", "init", root)
	out := svc.engineCLI(t, ctx, "", "plan", root)
	require.Contains(t, out, `+ foo = "bar"`)
	out = svc.engineCLI(t, ctx, "", "apply", root, "-auto-approve")
	require.Contains(t, out, `foo = "bar"`)

	// Edit variable and delete it
	browser.New(t, ctx, func(page playwright.Page) {
		//
		// go to workspace
		_, err = page.Goto(workspaceURL(svc.System.Hostname(), org.Name, "my-test-workspace"))
		require.NoError(t, err)
		// go to variables
		err = page.Locator(`//li[@id='menu-item-variables']/a`).Click()
		require.NoError(t, err)
		// edit variable
		err = page.Locator(`//tr[@id='item-variable-foo']//button[@id='edit-button']`).Click()
		require.NoError(t, err)
		// make it a 'sensitive' variable
		err = page.Locator("input#sensitive").Click()
		require.NoError(t, err)
		// submit form
		err = page.Locator(`//button[@id='save-variable-button']`).Click()
		require.NoError(t, err)

		// confirm variable updated
		err = expect.Locator(page.GetByRole("alert")).ToContainText("updated variable: foo")
		require.NoError(t, err)

		// confirm value is hidden (because it is sensitive)
		err = expect.Locator(page.Locator(`//*[@id='variables-table']//tbody/tr/td[3]/span[text()="hidden"]`)).ToBeVisible()
		require.NoError(t, err)
		// edit variable again
		err = page.Locator(`//tr[@id='item-variable-foo']//button[@id='edit-button']`).Click()
		require.NoError(t, err)

		// update value
		err = page.Locator("textarea#value").Fill("topsecret")
		require.NoError(t, err)
		screenshot(t, page, "variables_entering_top_secret")

		// submit form
		err = page.Locator(`//button[@id='save-variable-button']`).Click()
		require.NoError(t, err)

		// confirm variable updated
		err = expect.Locator(page.GetByRole("alert")).ToContainText("updated variable: foo")
		require.NoError(t, err)

		// delete variable
		err = page.Locator(`//button[@id='delete-button']`).Click()
		require.NoError(t, err)

		// confirm variable deleted
		err = expect.Locator(page.GetByRole("alert")).ToContainText("deleted variable: foo")
		require.NoError(t, err)
	})
}
