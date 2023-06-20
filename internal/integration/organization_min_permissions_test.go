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

	// Create org and its owner
	svc, org, ownerCtx := setup(t, nil)
	ws := svc.createWorkspace(t, ownerCtx, org)

	// Create user and add as member of guests team
	guest, guestCtx := svc.createUserCtx(t, ctx)
	guests := svc.createTeam(t, ownerCtx, org)
	err := svc.AddTeamMembership(ownerCtx, auth.TeamMembershipOptions{
		TeamID:   guests.ID,
		Username: guest.Username,
	})
	require.NoError(t, err)

	// Assign read role to guests team. Guests now receive a minimum set of
	// permissions across the workspace's organization.
	err = svc.SetPermission(ownerCtx, ws.ID, guests.Name, rbac.WorkspaceReadRole)
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
	provider := svc.createVCSProvider(t, ownerCtx, org)
	_, err = svc.GetVCSProvider(guestCtx, provider.ID)
	require.NoError(t, err)
}
