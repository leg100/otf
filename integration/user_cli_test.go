package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestUserCLI tests managing user accounts via the CLI
func TestUserCLI(t *testing.T) {
	t.Parallel()

	daemon := setup(t, nil)
	_, ctx := daemon.createUserCtx(t, ctx)

	// create user via cli
	out := daemon.otfcli(t, ctx, "users", "new", "bobby")
	require.Equal(t, "Successfully created user bobby\n", out)

	// delete user via cli
	out = daemon.otfcli(t, ctx, "users", "delete", "bobby")
	require.Equal(t, "Successfully deleted user bobby\n", out)
}
