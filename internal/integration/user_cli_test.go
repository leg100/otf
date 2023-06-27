package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestUserCLI tests managing user accounts via the CLI
func TestUserCLI(t *testing.T) {
	integrationTest(t)

	daemon, _, _ := setup(t, nil)

	// create user via cli
	out := daemon.otfcli(t, adminCtx, "users", "new", "bobby")
	require.Equal(t, "Successfully created user bobby\n", out)

	// delete user via cli
	out = daemon.otfcli(t, adminCtx, "users", "delete", "bobby")
	require.Equal(t, "Successfully deleted user bobby\n", out)
}
