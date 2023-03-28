package e2e

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestUserCLI tests managing user accounts via the CLI
func TestUserCLI(t *testing.T) {
	setup(t)

	daemon := &daemon{}
	daemon.withFlags("--site-token", "abc123")
	hostname := daemon.start(t)

	// create user
	cmd := exec.Command("otf", "users", "new", "bobby",
		"--address", hostname,
		"--token", "abc123",
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	require.Equal(t, "Successfully created user bobby\n", string(out))

	// delete user
	cmd = exec.Command("otf", "users", "delete", "bobby",
		"--address", hostname,
		"--token", "abc123",
	)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	require.Equal(t, "Successfully deleted user bobby\n", string(out))
}
