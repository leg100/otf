package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	daemon = "../_build/otfd"
	client = "../_build/otf"
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

func TestOTF(t *testing.T) {
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
		cmd := exec.Command(client, "organizations", "new", "automatize")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
	})

	t.Run("terraform init", func(t *testing.T) {
		chdir(t, root)
		cmd := exec.Command("terraform", "init", "-no-color")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
	})

	t.Run("terraform plan", func(t *testing.T) {
		chdir(t, root)
		cmd := exec.Command("terraform", "plan", "-no-color")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
	})

	t.Run("terraform apply", func(t *testing.T) {
		chdir(t, root)
		cmd := exec.Command("terraform", "apply", "-no-color", "-auto-approve")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
	})

	t.Run("lock workspace", func(t *testing.T) {
		cmd := exec.Command(client, "workspaces", "lock", "dev", "--organization", "automatize")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
	})

	t.Run("unlock workspace", func(t *testing.T) {
		cmd := exec.Command(client, "workspaces", "unlock", "dev", "--organization", "automatize")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
	})
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
