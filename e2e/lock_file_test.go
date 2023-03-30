package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

// TestLockFile tests various scenarios relating to the dependency lock file
// (.terraform.lock.hcl). Users may upload one, they may not, they may upload
// one with hashes for a different OS (e.g. Mac) and OTF is running on linux and
// has to then generates hashes for linux, etc. It is notorious for causing
// difficulties for users and it's no different for OTF.
func TestLockFile(t *testing.T) {
	org, workspace := setup(t)

	user := cloud.User{
		Name:  uuid.NewString(),
		Teams: []cloud.Team{{"owners", org}},
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
	})
	require.NoError(t, err)

	// create root module with only a variable and no resources - this should
	// result in *no* lock file being created.
	root := t.TempDir()
	config := []byte(fmt.Sprintf(`
terraform {
  backend "remote" {
	hostname = "%s"
	organization = "%s"

	workspaces {
	  name = "%s"
	}
  }
}

variable "foo" {
	default = "bar"
}
`, hostname, org, workspace))
	err = os.WriteFile(filepath.Join(root, "main.tf"), []byte(config), 0o600)
	require.NoError(t, err)

	// ensure tf cli has a token
	err = chromedp.Run(ctx, terraformLoginTasks(t, hostname))
	require.NoError(t, err)

	// verify terraform init and plan run without error
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
}
