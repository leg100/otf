package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestTeamCLI tests managing teams via the CLI
func TestTeamCLI(t *testing.T) {
	t.Parallel()

	daemon := setup(t, nil)
	_, ownerCtx := daemon.createUserCtx(t, ctx)

	// create organization
	out := daemon.otfcli(t, ownerCtx, "organizations", "new", "acme-corp")
	require.Equal(t, "Successfully created organization acme-corp\n", out)

	// create developers team
	out = daemon.otfcli(t, ownerCtx, "teams", "new", "devs", "--organization", "acme-corp")
	require.Equal(t, "Successfully created team devs\n", out)

	// create user via cli
	out = daemon.otfcli(t, ctx, "users", "new", "bobby")
	require.Equal(t, "Successfully created user bobby\n", out)

	// add user to developers
	out = daemon.otfcli(t, ctx, "teams", "add-user", "bobby",
		"--organization", "acme-corp",
		"--team", "devs",
	)
	require.Equal(t, "Successfully added bobby to devs\n", out)

	// remove user from team
	out = daemon.otfcli(t, ctx, "teams", "del-user", "bobby",
		"--organization", "acme-corp",
		"--team", "devs",
	)
	require.Equal(t, "Successfully removed bobby from devs\n", out)

	// delete team
	out = daemon.otfcli(t, ctx, "teams", "delete", "devs",
		"--organization", "acme-corp",
	)
	require.Equal(t, "Successfully deleted team devs\n", out)
}
