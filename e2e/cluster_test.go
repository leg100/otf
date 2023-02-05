package e2e

import (
	"os/exec"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCluster is an end-to-end test of the clustering capabilities, i.e.
// more than one otfd, both connected to the same postgres db. The test runs two
// otfd daemons:
//
// otfd1) otfd to which the TF CLI connects
// otfd2) otfd to which the otf-agent connects
//
// This setup provides a limited demonstration that the cluster is co-ordinating
// processes successfully, e.g. relaying of logs from the agent through to the
// TF CLI
func TestCluster(t *testing.T) {
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

	// start two daemons, one for user, one for agent
	daemon := &daemon{}
	daemon.withGithubUser(&user)
	userHostname := daemon.start(t)
	agentHostname := daemon.start(t)

	// create browser
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	// carry out browser tasks against user daemon
	err := chromedp.Run(ctx,
		// login to UI, which synchronises user's organization
		githubLoginTasks(t, userHostname, user.Name),
		// create and save API token
		terraformLoginTasks(t, userHostname),
	)
	require.NoError(t, err)

	// org now sync'd, so we can create agent token via CLI
	agentToken := createAgentToken(t, org, userHostname, "test-agent-token")
	// start agent, instructing it to connect to otfd2,
	// add --debug flag, which dumps info that test relies upon
	startAgent(t, agentToken, agentHostname, "--debug")

	// create root module, setting otfd1 as hostname
	root := newRootModule(t, userHostname, org, "dev")

	// terraform init automatically creates a workspace named dev
	cmd := exec.Command("terraform", "init", "-no-color")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)

	// edit workspace to use agent
	cmd = exec.Command("otf", "workspaces", "edit", "dev", "--organization", org, "--execution-mode", "agent", "--address", userHostname)
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	assert.Equal(t, "updated execution mode: agent\n", string(out))

	// terraform plan
	cmd = exec.Command("terraform", "plan", "-no-color")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	require.Contains(t, string(out), "Plan: 1 to add, 0 to change, 0 to destroy.")
	require.Contains(t, string(out), "External agent: true") // confirm run was handled by external agent

	// terraform apply
	cmd = exec.Command("terraform", "apply", "-no-color", "-auto-approve")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	require.Contains(t, string(out), "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.")
	require.Contains(t, string(out), "External agent: true") // confirm run was handled by external agent

	// delete agent token using UI
	okDialog(t, ctx)
	err = chromedp.Run(ctx, chromedp.Tasks{
		// go to org main menu
		chromedp.Navigate("https://" + userHostname + "/organizations/" + org),
		screenshot(t),
		// go to list of agent tokens
		chromedp.Click("#agent_tokens > a", chromedp.NodeVisible),
		screenshot(t),
		// delete the token
		chromedp.Click(`//button[text()='delete']`, chromedp.NodeVisible),
		screenshot(t),
		matchText(t, ".flash-success", "Deleted token: "+"test-agent-token"),
	})
	require.NoError(t, err)
}
