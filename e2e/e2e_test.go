package e2e

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOTF(t *testing.T) {
	addBuildsToPath(t)
	login(t)
	organization := createOrganization(t)
	root := newRoot(t, organization)

	// terraform init
	cmd := exec.Command("terraform", "init", "-no-color")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)

	// terraform plan
	cmd = exec.Command("terraform", "plan", "-no-color")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	require.Contains(t, string(out), "Plan: 1 to add, 0 to change, 0 to destroy.")

	// terraform apply
	cmd = exec.Command("terraform", "apply", "-no-color", "-auto-approve")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	require.Contains(t, string(out), "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.")

	// terraform destroy
	cmd = exec.Command("terraform", "destroy", "-no-color", "-auto-approve")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	require.NoError(t, err)
	t.Log(string(out))
	require.Contains(t, string(out), "Apply complete! Resources: 0 added, 0 changed, 1 destroyed.")

	// lock workspace
	cmd = exec.Command("otf", "workspaces", "lock", "dev", "--organization", organization)
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)

	// unlock workspace
	cmd = exec.Command("otf", "workspaces", "unlock", "dev", "--organization", organization)
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)

	// list workspaces
	cmd = exec.Command("otf", "workspaces", "list", "--organization", organization)
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	require.Contains(t, string(out), "dev")
}
