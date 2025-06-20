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

	svc, org, ctx := setup(t)
	ws := svc.createWorkspace(t, ctx, org)

	// Create user and add as member of guests team
	guest := svc.createUser(t)
	guests := svc.createTeam(t, ctx, org)
	err := svc.Users.AddTeamMembership(ctx, guests.ID, []user.Username{guest.Username})
	require.NoError(t, err)
	// Refresh guest user context to include new team membership
	_, guestCtx := svc.getUserCtx(t, adminCtx, guest.Username)

	// Assign read role to guests team. Guests now receive a minimum set of
	// permissions across the workspace's organization.
	err = svc.Workspaces.SetPermission(ctx, ws.ID, guests.ID, authz.WorkspaceReadRole)
	require.NoError(t, err)

	// Guest should be able to get org
	_, err = svc.Organizations.Get(guestCtx, org.Name)
	require.NoError(t, err)

	// Guest should be able to list teams
	_, err = svc.Teams.List(guestCtx, org.Name)
	require.NoError(t, err)

	// Guest should be able to list providers
	_, err = svc.VCSProviders.List(guestCtx, org.Name)
	require.NoError(t, err)

	// Guest should be able to get a provider
	provider := svc.createVCSProvider(t, ctx, org, nil)
	_, err = svc.VCSProviders.Get(guestCtx, provider.ID)
	require.NoError(t, err)
}
