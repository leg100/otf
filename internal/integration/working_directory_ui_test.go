package integration

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

// WorkingDirectory tests setting a working directory on a workspace and checks
// that terraform runs use configuration from that directory.
func TestWorkingDirectory(t *testing.T) {
	t.Parallel()

	daemon := setup(t, nil)
	user, ctx := daemon.createUserCtx(t, ctx)
	org := daemon.createOrganization(t, ctx)

	// create workspace and set working directory
	browser := createTabCtx(t)
	err := chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ctx, daemon.Hostname(), user.Username, daemon.Secret),
		createWorkspace(t, daemon.Hostname(), org.Name, "my-workspace"),
		chromedp.Tasks{
			// go to workspace
			chromedp.Navigate(workspaceURL(daemon.Hostname(), org.Name, "my-workspace")),
			screenshot(t),
			// go to workspace settings
			chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
			screenshot(t),
			// enter working directory
			chromedp.Focus("input#working_directory", chromedp.NodeVisible),
			input.InsertText("subdir"),
			screenshot(t),
			// submit form
			chromedp.Click(`//button[text()='Save changes']`, chromedp.NodeVisible),
			screenshot(t),
			// confirm workspace updated
			matchText(t, ".flash-success", "updated workspace"),
		},
	})
	require.NoError(t, err)

	// create root module along with a sub-directory containing the config we're
	// going to test
	root := newRootModule(t, daemon.Hostname(), org.Name, "my-workspace")
	subdir := path.Join(root, "subdir")
	err = os.Mkdir(subdir, 0o755)
	require.NoError(t, err)
	config := `
resource "null_resource" "subdir" {}
`
	err = os.WriteFile(filepath.Join(subdir, "main.tf"), []byte(config), 0o600)
	require.NoError(t, err)

	// run init in the *root* module
	_ = daemon.tfcli(t, ctx, "init", root)

	// run plan in the *root* module
	out := daemon.tfcli(t, ctx, "plan", root)
	require.Contains(t, string(out), `null_resource.subdir will be created`)

	// run apply in the *root* module
	out = daemon.tfcli(t, ctx, "apply", root, "-auto-approve")
	require.Contains(t, string(out), `null_resource.subdir: Creating...`)
	require.Contains(t, string(out), `null_resource.subdir: Creation complete`)
}
