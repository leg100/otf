package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestTeamCLI tests managing teams via the CLI
func TestTeamCLI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, nil)

	// create organization
	out := daemon.otfcli(t, ctx, "organizations", "new", "acme-corp")
	require.Equal(t, "Successfully created organization acme-corp\n", out)

	// create developers team
	out = daemon.otfcli(t, ctx, "teams", "new", "devs", "--organization", "acme-corp")
	require.Equal(t, "Successfully created team devs\n", out)

	// create users via cli
	out = daemon.otfcli(t, adminCtx, "users", "new", "bobby")
	require.Equal(t, "Successfully created user bobby\n", out)
	out = daemon.otfcli(t, adminCtx, "users", "new", "sally")
	require.Equal(t, "Successfully created user sally\n", out)

	// add users to developers
	out = daemon.otfcli(t, ctx, "teams", "add-users", "bobby", "sally",
		"--organization", "acme-corp",
		"--team", "devs",
	)
	require.Equal(t, "Successfully added [bobby sally] to devs\n", out)

	// remove users from team
	out = daemon.otfcli(t, ctx, "teams", "del-users", "bobby", "sally",
		"--organization", "acme-corp",
		"--team", "devs",
	)
	require.Equal(t, "Successfully removed [bobby sally] from devs\n", out)

	// delete team
	out = daemon.otfcli(t, ctx, "teams", "delete", "devs",
		"--organization", "acme-corp",
	)
	require.Equal(t, "Successfully deleted team devs\n", out)
}
