package e2e

import (
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	expect "github.com/google/goexpect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const agent = "../_build/otf-agent"

func TestAgent(t *testing.T) {
	tfpath := terraformPath(t)
	organization := createOrganization(t)
	root := newRoot(t, organization)

	t.Run("terraform login", func(t *testing.T) {
		login(t, root, tfpath)
	})

	t.Run("create organization", func(t *testing.T) {
		createOrganization(t)
	})

	t.Run("create agent", func(t *testing.T) {
		var token string
		t.Run("create agent token", func(t *testing.T) {
			cmd := exec.Command(client, "agents", "tokens", "new", "testing", "--organization", organization)
			out, err := cmd.CombinedOutput()
			t.Log(string(out))
			require.NoError(t, err)
			re := regexp.MustCompile(`Successfully created agent token: ([a-zA-Z0-9\-_]+)`)
			matches := re.FindStringSubmatch(string(out))
			require.Equal(t, 2, len(matches))
			token = matches[1]
		})

		// start agent process
		e, res, err := spawn(agent, "--token", token)
		require.NoError(t, err)

		out, err := e.ExpectBatch([]expect.Batcher{
			&expect.BExp{R: "successfully authenticated"},
			&expect.BExp{R: "stream update.*successfully connected"},
		}, time.Second*10)
		for _, o := range out {
			t.Log(o.Output)
		}
		require.NoError(t, err)

		// cleanly terminate agent at end of test
		t.Cleanup(func() {
			e.SendSignal(os.Interrupt)
			require.NoError(t, <-res)
		})

		// terraform init creates a workspace called dev
		t.Run("terraform init", func(t *testing.T) {
			chdir(t, root)
			cmd := exec.Command(tfpath, "init", "-no-color")
			out, err := cmd.CombinedOutput()
			t.Log(string(out))
			require.NoError(t, err)
		})

		t.Run("edit workspace to use agent", func(t *testing.T) {
			cmd := exec.Command(client, "workspaces", "edit", "dev", "--organization", organization, "--execution-mode", "agent")
			out, err := cmd.CombinedOutput()
			t.Log(string(out))
			require.NoError(t, err)
			assert.Equal(t, "updated execution mode: agent\n", string(out))
		})

		t.Run("terraform plan", func(t *testing.T) {
			chdir(t, root)
			cmd := exec.Command(tfpath, "plan", "-no-color")
			out, err := cmd.CombinedOutput()
			t.Log(string(out))
			require.NoError(t, err)
			require.Contains(t, string(out), "Plan: 1 to add, 0 to change, 0 to destroy.")
		})

		t.Run("terraform apply", func(t *testing.T) {
			chdir(t, root)
			cmd := exec.Command(tfpath, "apply", "-no-color", "-auto-approve")
			out, err := cmd.CombinedOutput()
			t.Log(string(out))
			require.NoError(t, err)
			require.Contains(t, string(out), "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.")
		})
	})
}

func spawn(command string, args ...string) (*expect.GExpect, <-chan error, error) {
	cmd := append([]string{command}, args...)
	return expect.SpawnWithArgs(cmd, time.Minute, expect.PartialMatch(true), expect.Verbose(testing.Verbose()))
}
