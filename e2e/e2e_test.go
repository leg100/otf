package e2e

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	daemon = "../_build/otsd"
	client = "../_build/ots"
	config = `
terraform {
  backend "remote" {
    hostname = "localhost:8080"
    organization = "automatize"

    workspaces {
      name = "dev"
    }
  }
}

resource "null_resource" "e2e" {}
`
	rc = `
credentials "localhost:8080" {
  token = "xxxxxx.atlasv1.zzzzzzzzzzzzz"
}
`
)

func TestOTS(t *testing.T) {
	tmpdir := t.TempDir()
	dbPath := filepath.Join(tmpdir, "e2e.db")
	logFile, err := os.Create("e2e.log")
	require.NoError(t, err)

	// Write TF config file
	rcPath := filepath.Join(tmpdir, ".terraformrc")
	require.NoError(t, os.WriteFile(rcPath, []byte(rc), 0600))
	os.Setenv("TF_CLI_CONFIG_FILE", rcPath)

	// Run OTS daemon
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, daemon, "-db-path", dbPath)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	defer cancel()
	require.NoError(t, cmd.Start())
	wait(t)

	// Create org
	createOrg(t, logFile)

	// Create TF config
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "main.tf"), []byte(config), 0600))

	t.Run("terraform init", func(t *testing.T) {
		chdir(t, root)
		cmd = exec.Command("terraform", "init", "-no-color")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err)
		t.Log(string(out))
	})

	t.Run("terraform plan", func(t *testing.T) {
		chdir(t, root)
		cmd = exec.Command("terraform", "plan", "-no-color")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err)
		t.Log(string(out))
	})

	t.Run("terraform apply", func(t *testing.T) {
		chdir(t, root)
		cmd = exec.Command("terraform", "apply", "-no-color", "-auto-approve")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err)
		t.Log(string(out))
	})
}

// Wait for OTS to start
func wait(t *testing.T) {
	// Ping daemon five times, with a one second interval
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)

		req, err := http.NewRequest("GET", "https://localhost:8080/api/v2/ping", nil)
		req.Header.Add("Accept", "application/vnd.api+json")
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			if resp.StatusCode == 204 {
				return
			}
			t.Logf("received status code: %d", resp.StatusCode)
		}
		t.Logf("received error: %s", err.Error())
	}
	t.Error("daemon failed to start")
}

// Seed DB with organization
func createOrg(t *testing.T, logFile *os.File) {
	cmd := exec.Command(client, "organizations", "new", "automatize", "--email", "e2e@automatize.co")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	require.NoError(t, cmd.Run())
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
