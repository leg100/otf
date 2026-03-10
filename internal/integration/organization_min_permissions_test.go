package integration

import (
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/require"
)

// TestIntegration_MinimumPermissions demonstrates that once a user is
// indirectly a member of an organization - i.e. they have been assigned at least
// a role on a workspace in the organization - that they receive a minimum set
// of permissions across the organization.
func TestIntegration_MinimumPermissions(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)
	ws := daemon.createWorkspace(t, ctx, org)

	// Create user and add as member of guests team
	guest := daemon.createUser(t)
	guests := daemon.createTeam(t, ctx, org)
	err := daemon.Users.AddTeamMembership(ctx, guests.ID, []user.Username{guest.Username})
	require.NoError(t, err)
	// Refresh guest user context to include new team membership
	_, guestCtx := daemon.getUserCtx(t, adminCtx, guest.Username)

	// Assign read role to guests team. Guests now receive a minimum set of
	// permissions across the workspace's organization.
	err = daemon.Workspaces.SetWorkspacePermission(ctx, ws.ID, guests.ID, authz.WorkspaceReadRole)
	require.NoError(t, err)

	// Guest should be able to get org
	_, err = daemon.Organizations.GetOrganization(guestCtx, org.Name)
	require.NoError(t, err)

	// Guest should be able to list teams
	_, err = daemon.Teams.ListTeams(guestCtx, org.Name)
	require.NoError(t, err)

	// Guest should be able to list providers
	_, err = daemon.VCSProviders.ListVCSProviders(guestCtx, org.Name)
	require.NoError(t, err)

	// Guest should be able to get a provider
	provider := daemon.createVCSProvider(t, ctx, org, nil)
	_, err = daemon.VCSProviders.GetVCSProvider(guestCtx, provider.ID)
	require.NoError(t, err)
}
