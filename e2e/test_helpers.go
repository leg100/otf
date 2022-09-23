package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	expect "github.com/google/goexpect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	daemon = "../_build/otfd"
	client = "../_build/otf"
	agent  = "../_build/otf-agent"
)

func startDaemon(t *testing.T, port int) {
	e, res, err := spawn(daemon,
		"--address",
		fmt.Sprintf(":%d", port),
		"--cert-file", "./fixtures/cert.crt",
		"--key-file", "./fixtures/key.pem",
	)
	require.NoError(t, err)

	_, _, err = e.Expect(regexp.MustCompile("started server"), time.Second*10)
	require.NoError(t, err)

	t.Cleanup(func() {
		e.SendSignal(os.Interrupt)
		require.NoError(t, <-res)
	})
}

func startAgent(t *testing.T, token, address string) {
	out, err := os.CreateTemp(t.TempDir(), "agent.out")
	require.NoError(t, err)

	e, res, err := expect.Spawn(
		fmt.Sprintf("%s --token %s --address %s", agent, token, address),
		time.Minute,
		expect.PartialMatch(true),
		expect.Verbose(testing.Verbose()),
		expect.Tee(out),
	)
	require.NoError(t, err)

	_, err = e.ExpectBatch([]expect.Batcher{
		&expect.BExp{R: "successfully authenticated"},
		&expect.BExp{R: "stream update.*successfully connected"},
	}, time.Second*10)
	require.NoError(t, err)

	// terminate at end of parent test
	t.Cleanup(func() {
		e.SendSignal(os.Interrupt)
		if !assert.NoError(t, <-res) || t.Failed() {
			logs, err := os.ReadFile(out.Name())
			require.NoError(t, err)
			t.Log("--- agent logs ---")
			t.Log(string(logs))
		}
	})
}

func createAgentToken(t *testing.T, organization string) string {
	cmd := exec.Command(client, "agents", "tokens", "new", "testing", "--organization", organization)
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	re := regexp.MustCompile(`Successfully created agent token: (agent\.[a-zA-Z0-9\-_]+)`)
	matches := re.FindStringSubmatch(string(out))
	require.Equal(t, 2, len(matches))
	return matches[1]
}

func spawn(command string, args ...string) (*expect.GExpect, <-chan error, error) {
	cmd := append([]string{command}, args...)
	return expect.SpawnWithArgs(cmd, time.Minute, expect.PartialMatch(true), expect.Verbose(testing.Verbose()))
}

// Chdir changes current directory to this temp directory.
func chdir(t *testing.T, dir string) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal("unable to get current directory")
	}

	t.Cleanup(func() {
		if err := os.Chdir(pwd); err != nil {
			t.Fatal("unable to reset current directory")
		}
	})

	if err := os.Chdir(dir); err != nil {
		t.Fatal("unable to change current directory")
	}
}

// newRoot creates a root module
func newRoot(t *testing.T, organization string) string {
	config := []byte(fmt.Sprintf(`
terraform {
  backend "remote" {
	hostname = "localhost:8080"
	organization = "%s"

	workspaces {
	  name = "dev"
	}
  }
}
resource "null_resource" "e2e" {}
`, organization))

	root := t.TempDir()
	err := os.WriteFile(filepath.Join(root, "main.tf"), config, 0o600)
	require.NoError(t, err)

	return root
}

func createOrganization(t *testing.T) string {
	organization := uuid.NewString()
	cmd := exec.Command(client, "organizations", "new", organization)
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	return organization
}

func login(t *testing.T, tfpath string) func(t *testing.T) {
	return func(t *testing.T) {
		// nullifying PATH ensures `terraform login` skips opening a browser
		// window
		t.Setenv("PATH", "")

		token, foundToken := os.LookupEnv("OTF_SITE_TOKEN")
		if !foundToken {
			t.Fatal("Test cannot proceed without OTF_SITE_TOKEN")
		}

		e, tferr, err := expect.SpawnWithArgs(
			[]string{tfpath, "login", "localhost:8080"},
			time.Minute,
			expect.PartialMatch(true),
			expect.Verbose(testing.Verbose()))
		require.NoError(t, err)
		defer e.Close()

		e.ExpectBatch([]expect.Batcher{
			&expect.BExp{R: "Enter a value:"}, &expect.BSnd{S: "yes\n"},
			&expect.BExp{R: "Enter a value:"}, &expect.BSnd{S: token + "\n"},
			&expect.BExp{R: "Success! Logged in to Terraform Enterprise"},
		}, time.Minute)
		require.NoError(t, <-tferr)
	}
}

func terraformPath(t *testing.T) string {
	path, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal("terraform executable not found in path")
	}
	return path
}
