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
	t.Parallel()

	svc, org, ctx := setup(t, nil)

	// Create variable in browser
	browser.Run(t, ctx, chromedp.Tasks{
		createWorkspace(t, svc.Hostname(), org.Name, "my-test-workspace"),
		chromedp.Tasks{
			// go to workspace
			chromedp.Navigate(workspaceURL(svc.Hostname(), org.Name, "my-test-workspace")),
			screenshot(t),
			// go to variables
			chromedp.Click(`//a[text()='variables']`, chromedp.NodeVisible),
			screenshot(t),
			// click add variable button
			chromedp.Click(`//button[text()='Add variable']`, chromedp.NodeVisible),
			screenshot(t),
			// enter key
			chromedp.Focus("input#key", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("foo"),
			screenshot(t),
			// enter value
			chromedp.Focus("textarea#value", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("bar"),
			screenshot(t),
			// select terraform variable category
			chromedp.Click("input#terraform", chromedp.NodeVisible, chromedp.ByQuery),
			screenshot(t),
			// submit form
			chromedp.Click(`//button[@id='save-variable-button']`, chromedp.NodeVisible),
			screenshot(t),
			// confirm variable added
			matchText(t, ".flash-success", "added variable: foo"),
			screenshot(t),
		},
	})

	// write some terraform config that declares and outputs the variable
	root := newRootModule(t, svc.Hostname(), org.Name, "my-test-workspace")
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
	browser.Run(t, ctx, chromedp.Tasks{
		chromedp.Tasks{
			// go to workspace
			chromedp.Navigate(workspaceURL(svc.Hostname(), org.Name, "my-test-workspace")),
			screenshot(t),
			// go to variables
			chromedp.Click(`//a[text()='variables']`, chromedp.NodeVisible),
			screenshot(t),
			// edit variable
			chromedp.Click(`//a[text()='foo']`, chromedp.NodeVisible),
			screenshot(t),
			// make it a 'sensitive' variable
			chromedp.Click("input#sensitive", chromedp.NodeVisible, chromedp.ByQuery),
			screenshot(t),
			// submit form
			chromedp.Click(`//button[@id='save-variable-button']`, chromedp.NodeVisible),
			screenshot(t),
			// confirm variable updated
			matchText(t, ".flash-success", "updated variable: foo"),
			screenshot(t),
			// confirm value is hidden (because it is sensitive)
			matchText(t, `//table[@class='variables']/tbody/tr/td[2]`, "hidden"),
			// edit variable again
			chromedp.Click(`//a[text()='foo']`, chromedp.NodeVisible),
			screenshot(t),
			// update value
			chromedp.Focus("textarea#value", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("topsecret"),
			screenshot(t, "variables_entering_top_secret"),
			// submit form
			chromedp.Click(`//button[@id='save-variable-button']`, chromedp.NodeVisible),
			screenshot(t),
			// confirm variable updated
			matchText(t, ".flash-success", "updated variable: foo"),
			screenshot(t),
			// delete variable
			chromedp.Click(`//button[@id='delete-variable-button']`, chromedp.NodeVisible),
			screenshot(t),
			// confirm variable deleted
			matchText(t, ".flash-success", "deleted variable: foo"),
		},
	})
}
