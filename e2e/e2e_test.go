package e2e

import (
	"bytes"
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
)

func TestOTS(t *testing.T) {
	tmpdir := t.TempDir()
	dbPath := filepath.Join(tmpdir, "e2e.db")

	// Run OTS daemon
	out := new(bytes.Buffer)
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, daemon, "--db-path", dbPath, "--ssl", "--cert-file", "fixtures/cert.crt", "--key-file", "fixtures/key.pem")
	cmd.Stdout = out
	cmd.Stderr = out
	defer func() {
		t.Log("--- daemon output ---")
		t.Log(out.String())
	}()
	defer cancel()
	require.NoError(t, cmd.Start())
	wait(t)

	// Create TF config
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "main.tf"), []byte(config), 0600))

	t.Run("login", func(t *testing.T) {
		cmd := exec.Command(client, "login")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
	})

	t.Run("create organization", func(t *testing.T) {
		cmd := exec.Command(client, "organizations", "new", "automatize", "--email", "e2e@automatize.co")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
	})

	t.Run("terraform init", func(t *testing.T) {
		chdir(t, root)
		cmd = exec.Command("terraform", "init", "-no-color")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
	})

	t.Run("terraform plan", func(t *testing.T) {
		chdir(t, root)
		cmd = exec.Command("terraform", "plan", "-no-color")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
	})

	t.Run("terraform apply", func(t *testing.T) {
		chdir(t, root)
		cmd = exec.Command("terraform", "apply", "-no-color", "-auto-approve")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
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
