package e2e

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

// WorkingDirectory tests setting a working directory on a workspace and checks
// that terraform runs use configuration from that directory.
func TestWorkingDirectory(t *testing.T) {
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

	// login, create workspace and set working directory
	err := chromedp.Run(ctx, chromedp.Tasks{
		githubLoginTasks(t, hostname, user.Name),
		createWorkspaceTasks(t, hostname, org, workspace),
		chromedp.Tasks{
			// go to workspace
			chromedp.Navigate(workspacePath(hostname, org, workspace)),
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
	root := newRootModule(t, hostname, org, workspace)
	subdir := path.Join(root, "subdir")
	err = os.Mkdir(subdir, 0o755)
	require.NoError(t, err)
	config := `
resource "null_resource" "subdir" {}
`
	err = os.WriteFile(filepath.Join(subdir, "main.tf"), []byte(config), 0o600)
	require.NoError(t, err)

	// ensure tf cli has a token
	err = chromedp.Run(ctx, terraformLoginTasks(t, hostname))
	require.NoError(t, err)

	// run init in the *root* module
	cmd := exec.Command("terraform", "init", "-no-color")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)

	// run plan in the *root* module
	cmd = exec.Command("terraform", "plan", "-no-color")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	require.Contains(t, string(out), `null_resource.subdir will be created`)

	// run apply in the *root* module
	cmd = exec.Command("terraform", "apply", "-no-color", "-auto-approve")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	require.Contains(t, string(out), `null_resource.subdir: Creating...`)
	require.Contains(t, string(out), `null_resource.subdir: Creation complete`)
}
