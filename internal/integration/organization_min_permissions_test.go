package integration

import (
	"testing"

	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/rbac"
	"github.com/stretchr/testify/require"
)

// TestIntegration_MinimumPermissions demonstrates that once a user is
// indirectly a member of an organization - i.e. they have been assigned at least
// a role on a workspace in the organization - that they receive a minimum set
// of permissions across the organization.
func TestIntegration_MinimumPermissions(t *testing.T) {
	t.Parallel()

	svc, org, ctx := setup(t, nil)
	ws := svc.createWorkspace(t, ctx, org)

	// Create user and add as member of guests team
	guest := svc.createUser(t)
	guests := svc.createTeam(t, ctx, org)
	err := svc.AddTeamMembership(ctx, auth.TeamMembershipOptions{
		TeamID:   guests.ID,
		Username: guest.Username,
	})
	require.NoError(t, err)
	// Refresh guest user context to include new team membership
	_, guestCtx := svc.getUserCtx(t, adminCtx, guest.Username)

	// Assign read role to guests team. Guests now receive a minimum set of
	// permissions across the workspace's organization.
	err = svc.SetPermission(ctx, ws.ID, guests.Name, rbac.WorkspaceReadRole)
	require.NoError(t, err)

	// Guest should be able to get org
	_, err = svc.GetOrganization(guestCtx, org.Name)
	require.NoError(t, err)

	// Guest should be able to list teams
	_, err = svc.ListTeams(guestCtx, org.Name)
	require.NoError(t, err)

	// Guest should be able to list providers
	_, err = svc.ListVCSProviders(guestCtx, org.Name)
	require.NoError(t, err)

	// Guest should be able to get a provider
	provider := svc.createVCSProvider(t, ctx, org)
	_, err = svc.GetVCSProvider(guestCtx, provider.ID)
	require.NoError(t, err)
}
