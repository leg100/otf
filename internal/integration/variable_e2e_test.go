package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

// TestVariableE2E tests adding, updating and deleting workspace variables via the
// UI, and tests that variables are made available to runs.
func TestVariableE2E(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	// Create variable in browser
	page := browser.New(t, ctx)
		createWorkspace(t, svc.System.Hostname(), org.Name, "my-test-workspace"),
		chromedp.Tasks{
			// go to workspace
			_, err = page.Goto(workspaceURL(svc.System.Hostname(), org.Name, "my-test-workspace"))
require.NoError(t, err)
			//screenshot(t),
			// go to variables
			err := page.Locator(`//a[text()='variables']`).Click()
require.NoError(t, err)
			//screenshot(t),
			// click add variable button
			err := page.Locator(`//button[text()='Add variable']`).Click()
require.NoError(t, err)
			//screenshot(t),
			// enter key
			chromedp.Focus("input#key", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("foo"),
			//screenshot(t),
			// enter value
			chromedp.Focus("textarea#value", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("bar"),
			//screenshot(t),
			// select terraform variable category
			err := page.Locator("input#terraform").Click()
require.NoError(t, err)
			//screenshot(t),
			// submit form
			err := page.Locator(`//button[@id='save-variable-button']`).Click()
require.NoError(t, err)
			//screenshot(t),
			// confirm variable added
			matchText(t, "//div[@role='alert']", "added variable: foo"),
			//screenshot(t),
		},
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
	svc.tfcli(t, ctx, "init", root)
	out := svc.tfcli(t, ctx, "plan", root)
	require.Contains(t, out, `+ foo = "bar"`)
	out = svc.tfcli(t, ctx, "apply", root, "-auto-approve")
	require.Contains(t, out, `foo = "bar"`)

	// Edit variable and delete it
	page := browser.New(t, ctx)
		chromedp.Tasks{
			// go to workspace
			_, err = page.Goto(workspaceURL(svc.System.Hostname(), org.Name, "my-test-workspace"))
require.NoError(t, err)
			//screenshot(t),
			// go to variables
			err := page.Locator(`//a[text()='variables']`).Click()
require.NoError(t, err)
			//screenshot(t),
			// edit variable
			err := page.Locator(`//a[text()='foo']`).Click()
require.NoError(t, err)
			//screenshot(t),
			// make it a 'sensitive' variable
			err := page.Locator("input#sensitive").Click()
require.NoError(t, err)
			//screenshot(t),
			// submit form
			err := page.Locator(`//button[@id='save-variable-button']`).Click()
require.NoError(t, err)
			//screenshot(t),
			// confirm variable updated
			chromedp.WaitVisible(`//div[@role='alert'][contains(text(),"updated variable: foo")]`),
			//screenshot(t),
			// confirm value is hidden (because it is sensitive)
			chromedp.WaitVisible(`//table[@id='variables-table']/tbody/tr/td[2]/span[text()="hidden"]`),
			// edit variable again
			err := page.Locator(`//a[text()='foo']`).Click()
require.NoError(t, err)
			//screenshot(t),
			// update value
			chromedp.Focus("textarea#value", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("topsecret"),
			//screenshot(t, "variables_entering_top_secret"),
			// submit form
			err := page.Locator(`//button[@id='save-variable-button']`).Click()
require.NoError(t, err)
			//screenshot(t),
			// confirm variable updated
			chromedp.WaitVisible(`//div[@role='alert'][contains(text(),"updated variable: foo")]`),
			//screenshot(t),
			// delete variable
			err := page.Locator(`//button[@id='delete-variable-button']`).Click()
require.NoError(t, err)
			//screenshot(t),
			// confirm variable deleted
			chromedp.WaitVisible(`//div[@role='alert'][contains(text(),"deleted variable: foo")]`),
		},
	})
}
