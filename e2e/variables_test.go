package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

// TestVariables tests adding, updating and deleting workspace variables via the
// UI, and tests that variables are made available to runs.
func TestVariables(t *testing.T) {
	org, workspace := setup(t)

	user := cloud.User{
		Name:  uuid.NewString(),
		Teams: []cloud.Team{{Name: "owners", Organization: org}},
	}

	daemon := &daemon{}
	daemon.withGithubUser(&user)
	hostname := daemon.start(t)

	// create browser
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	// Click OK on any browser javascript dialog boxes that pop up
	okDialog(t, ctx)

	// Create variable in browser
	err := chromedp.Run(ctx, chromedp.Tasks{
		githubLoginTasks(t, hostname, user.Name),
		createWorkspaceTasks(t, hostname, org, workspace),
		chromedp.Tasks{
			// go to workspace
			chromedp.Navigate(workspacePath(hostname, org, workspace)),
			screenshot(t),
			// go to variables
			chromedp.Click(`//a[text()='variables']`, chromedp.NodeVisible),
			screenshot(t),
			// click add variable button
			chromedp.Click(`//button[text()='Add variable']`, chromedp.NodeVisible),
			screenshot(t),
			// enter key
			chromedp.Focus("input#key", chromedp.NodeVisible),
			input.InsertText("foo"),
			screenshot(t),
			// enter value
			chromedp.Focus("textarea#value", chromedp.NodeVisible),
			input.InsertText("bar"),
			screenshot(t),
			// select terraform variable category
			chromedp.Click("input#terraform", chromedp.NodeVisible),
			screenshot(t),
			// submit form
			chromedp.Click(`//button[text()='Save variable']`, chromedp.NodeVisible),
			screenshot(t),
			// confirm variable added
			matchText(t, ".flash-success", "added variable: foo"),
			screenshot(t),
		},
	})
	require.NoError(t, err)

	// write some terraform config that declares and outputs the variable
	root := newRootModule(t, hostname, org, workspace)
	config := `
variable "foo" {
  default = "overwrite_this"
}

output "foo" {
  value = var.foo
}
`
	err = os.WriteFile(filepath.Join(root, "foo.tf"), []byte(config), 0o600)
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
	require.Contains(t, string(out), `+ foo = "bar"`)

	cmd = exec.Command("terraform", "apply", "-no-color", "-auto-approve")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	require.Contains(t, string(out), `foo = "bar"`)

	// Edit variable and delete it
	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Tasks{
			// go to workspace
			chromedp.Navigate(workspacePath(hostname, org, workspace)),
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
			chromedp.Click(`//button[text()='Save variable']`, chromedp.NodeVisible),
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
			chromedp.Focus("textarea#value", chromedp.NodeVisible),
			input.InsertText("topsecret"),
			screenshot(t),
			// submit form
			chromedp.Click(`//button[text()='Save variable']`, chromedp.NodeVisible),
			screenshot(t),
			// confirm variable updated
			matchText(t, ".flash-success", "updated variable: foo"),
			screenshot(t),
			// delete variable
			chromedp.Click(`//button[text()='Delete']`, chromedp.NodeVisible),
			screenshot(t),
			// confirm variable deleted
			matchText(t, ".flash-success", "deleted variable: foo"),
		},
	})
	require.NoError(t, err)
}
