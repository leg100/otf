package e2e

import (
	"os/exec"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

// TestStartRunUI tests starting a run via the Web UI before confirming and
// applying the run.
func TestStartRunUI(t *testing.T) {
	addBuildsToPath(t)

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
	daemon := &daemon{}
	daemon.withGithubUser(&user)
	hostname := daemon.start(t)
	workspaceName := "workspace-start-run"
	root := newRootModule(t, hostname, org, workspaceName)

	// create browser
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	// in browser, login and create workspace
	err := chromedp.Run(ctx,
		githubLoginTasks(t, hostname, user.Name),
		createWorkspaceTasks(t, hostname, org, workspaceName),
		terraformLoginTasks(t, hostname),
	)
	require.NoError(t, err)

	//
	// start run UI functionality requires an existing config version, so
	// create one first by running a plan via the CLI
	//

	// terraform init
	cmd := exec.Command("terraform", "init", "-no-color")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)

	// terraform plan
	cmd = exec.Command("terraform", "plan", "-no-color")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	require.Contains(t, string(out), "Plan: 1 to add, 0 to change, 0 to destroy.")

	// now we have a config version, start a run via the browser
	err = chromedp.Run(ctx, startRunTasks(t, hostname, org, workspaceName))
	require.NoError(t, err)
}
