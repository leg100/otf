package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	expect "github.com/google/goexpect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	daemon = "../_build/otfd"
	client = "../_build/otf"
)

func newConfig(organization string) []byte {
	config := `
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
`
	return []byte(fmt.Sprintf(config, organization))
}

func TestOTF(t *testing.T) {
	tfpath, err := exec.LookPath("terraform")
	require.NoError(t, err)

	// Create TF config
	root := t.TempDir()
	organization := uuid.NewString()
	err = os.WriteFile(filepath.Join(root, "main.tf"), newConfig(organization), 0o600)
	require.NoError(t, err)

	t.Run("terraform login", func(t *testing.T) {
		// nullifying PATH ensures `terraform login` skips opening a browser
		// window
		t.Setenv("PATH", "")

		token, foundToken := os.LookupEnv("OTF_SITE_TOKEN")
		if !foundToken {
			t.Fatal("Test cannot proceed without OTF_SITE_TOKEN")
		}

		chdir(t, root)

		e, tferr, err := expect.Spawn(fmt.Sprintf("%s login localhost:8080", tfpath), time.Minute, expect.PartialMatch(true), expect.Verbose(testing.Verbose()))
		require.NoError(t, err)
		defer e.Close()

		e.ExpectBatch([]expect.Batcher{
			&expect.BExp{R: "Enter a value:"}, &expect.BSnd{S: "yes\n"},
			&expect.BExp{R: "Enter a value:"}, &expect.BSnd{S: token + "\n"},
			&expect.BExp{R: "Success! Logged in to Terraform Enterprise"},
		}, time.Minute)
		require.NoError(t, <-tferr)
	})

	t.Run("create organization", func(t *testing.T) {
		cmd := exec.Command(client, "organizations", "new", organization)
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
	})

	t.Run("terraform init", func(t *testing.T) {
		chdir(t, root)
		cmd := exec.Command(tfpath, "init", "-no-color")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
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

	t.Run("terraform destroy", func(t *testing.T) {
		chdir(t, root)
		cmd := exec.Command(tfpath, "destroy", "-no-color", "-auto-approve")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err)
		t.Log(string(out))
		require.Contains(t, string(out), "Apply complete! Resources: 0 added, 0 changed, 1 destroyed.")
	})

	t.Run("lock workspace", func(t *testing.T) {
		cmd := exec.Command(client, "workspaces", "lock", "dev", "--organization", organization)
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
	})

	t.Run("unlock workspace", func(t *testing.T) {
		cmd := exec.Command(client, "workspaces", "unlock", "dev", "--organization", organization)
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
	})

	t.Run("list workspaces", func(t *testing.T) {
		cmd := exec.Command(client, "workspaces", "list", "--organization", organization)
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
		require.Contains(t, string(out), "dev")
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
