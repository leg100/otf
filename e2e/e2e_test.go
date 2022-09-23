package e2e

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOTF(t *testing.T) {
	tfpath := terraformPath(t)
	t.Run("login", login(t, tfpath))
	organization := createOrganization(t)
	root := newRoot(t, organization)

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
