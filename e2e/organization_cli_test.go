package e2e

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestOrganizationCLI tests managing organizations via the CLI
func TestOrganizationCLI(t *testing.T) {
	setup(t)

	daemon := &daemon{}
	daemon.withFlags("--site-token", "abc123")
	hostname := daemon.start(t)

	// create organization
	cmd := exec.Command("otf", "organizations", "new", "acme-corp",
		"--address", hostname,
		"--token", "abc123",
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	require.Equal(t, "Successfully created organization acme-corp\n", string(out))

	// create user
	cmd = exec.Command("otf", "users", "new", "bobby",
		"--address", hostname,
		"--token", "abc123",
	)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	require.Equal(t, "Successfully created user bobby\n", string(out))

	// add user to organization
	cmd = exec.Command("otf", "organizations", "add-user", "bobby",
		"--address", hostname,
		"--token", "abc123",
		"--organization", "acme-corp",
	)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	require.Equal(t, "Successfully added bobby to acme-corp\n", string(out))

	// remove user from organization
	cmd = exec.Command("otf", "organizations", "del-user", "bobby",
		"--address", hostname,
		"--token", "abc123",
		"--organization", "acme-corp",
	)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	require.Equal(t, "Successfully removed bobby from acme-corp\n", string(out))

	// delete organization
	cmd = exec.Command("otf", "organizations", "delete", "acme-corp",
		"--address", hostname,
		"--token", "abc123",
	)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	require.Equal(t, "Successfully deleted organization acme-corp\n", string(out))
}
