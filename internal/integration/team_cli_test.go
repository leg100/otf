package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestTeamCLI tests managing teams via the CLI
func TestTeamCLI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t)

	// create organization
	out := daemon.otfCLI(t, ctx, "organizations", "new", "acme-corp")
	require.Equal(t, "Successfully created organization acme-corp\n", out)

	// create developers team
	out = daemon.otfCLI(t, ctx, "teams", "new", "devs", "--organization", "acme-corp")
	require.Equal(t, "Successfully created team devs\n", out)

	// create users via cli
	out = daemon.otfCLI(t, adminCtx, "users", "new", "bobby")
	require.Equal(t, "Successfully created user bobby\n", out)
	out = daemon.otfCLI(t, adminCtx, "users", "new", "sally")
	require.Equal(t, "Successfully created user sally\n", out)

	// add users to developers
	out = daemon.otfCLI(t, ctx, "team-membership", "add-users", "bobby", "sally",
		"--organization", "acme-corp",
		"--team", "devs",
	)
	require.Equal(t, "Successfully added [bobby sally] to devs\n", out)

	// remove users from team
	out = daemon.otfCLI(t, ctx, "team-membership", "del-users", "bobby", "sally",
		"--organization", "acme-corp",
		"--team", "devs",
	)
	require.Equal(t, "Successfully removed [bobby sally] from devs\n", out)

	// delete team
	out = daemon.otfCLI(t, ctx, "teams", "delete", "devs",
		"--organization", "acme-corp",
	)
	require.Equal(t, "Successfully deleted team devs\n", out)
}
