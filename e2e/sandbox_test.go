package e2e

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSandbox demonstrates the sandbox feature, whereby terraform apply is run
// within an isolated environment.
func TestSandbox(t *testing.T) {
	if _, err := exec.LookPath("bwrap"); errors.Is(err, exec.ErrNotFound) {
		t.Skipf("bwrap binary not found")
	}

	addBuildsToPath(t)

	org := uuid.NewString()
	user := cloud.User{
		Name: "cluster-user",
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
	daemon.withFlags("--sandbox")
	hostname := daemon.start(t)

	// create terraform config
	config := newRootModule(t, hostname, org, "dev")

	// create browser
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	// create api token and pass token to terraform login
	err := chromedp.Run(ctx, chromedp.Tasks{
		githubLoginTasks(t, hostname, user.Name),
		terraformLoginTasks(t, hostname),
	})
	require.NoError(t, err)

	cmd := exec.Command("terraform", "init", "-no-color")
	cmd.Dir = config
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)

	// terraform apply
	cmd = exec.Command("terraform", "apply", "-no-color", "-auto-approve")
	cmd.Dir = config
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	assert.Contains(t, string(out), "Running within sandbox...")
	assert.Contains(t, string(out), "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.")
}
