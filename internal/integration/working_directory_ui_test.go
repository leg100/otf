package integration

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// WorkingDirectory tests setting a working directory on a workspace and checks
// that terraform runs use configuration from that directory.
func TestWorkingDirectory(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)

	// create workspace and set working directory
	page := browser.New(t, ctx)

	createWorkspace(t, page, daemon.System.Hostname(), org.Name, "my-workspace")

	// go to workspace
	_, err := page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, "my-workspace"))
	require.NoError(t, err)
	//screenshot(t),

	// go to workspace settings
	err = page.Locator(`//a[text()='settings']`).Click()
	require.NoError(t, err)
	//screenshot(t),

	// enter working directory
	err = page.Locator("input#working_directory").Fill("subdir")
	require.NoError(t, err)
	//screenshot(t),

	// submit form
	err = page.Locator(`//button[text()='Save changes']`).Click()
	require.NoError(t, err)
	//screenshot(t),
	// confirm workspace updated
	err = expect.Locator(page.Locator("//div[@role='alert']")).ToHaveText("updated workspace")
	require.NoError(t, err)

	// create root module along with a sub-directory containing the config we're
	// going to test
	root := newRootModule(t, daemon.System.Hostname(), org.Name, "my-workspace")
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
