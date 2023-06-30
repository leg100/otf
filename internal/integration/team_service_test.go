package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegation_TeamService(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		svc, org, ctx := setup(t, nil)

		team, err := svc.CreateTeam(ctx, auth.CreateTeamOptions{
			Name:         uuid.NewString(),
			Organization: org.Name,
		})
		require.NoError(t, err)

		t.Run("already exists error", func(t *testing.T) {
			_, err := svc.CreateTeam(ctx, auth.CreateTeamOptions{
				Name:         team.Name,
				Organization: org.Name,
			})
			require.Equal(t, internal.ErrResourceAlreadyExists, err)
		})
	})

	t.Run("update", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		team := svc.createTeam(t, ctx, nil)

		_, err := svc.UpdateTeam(ctx, team.ID, auth.UpdateTeamOptions{
			OrganizationAccessOptions: auth.OrganizationAccessOptions{
				ManageWorkspaces: internal.Bool(true),
				ManageVCS:        internal.Bool(true),
				ManageModules:    internal.Bool(true),
			},
		})
		require.NoError(t, err)

		got, err := svc.GetTeam(ctx, team.Organization, team.Name)
		require.NoError(t, err)

		assert.True(t, got.OrganizationAccess().ManageWorkspaces)
		assert.True(t, got.OrganizationAccess().ManageVCS)
		assert.True(t, got.OrganizationAccess().ManageModules)
	})

	t.Run("get", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		team := svc.createTeam(t, ctx, nil)

		got, err := svc.GetTeam(ctx, team.Organization, team.Name)
		require.NoError(t, err)

		assert.Equal(t, team, got)
	})

	t.Run("get by id", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		want := svc.createTeam(t, ctx, nil)

		got, err := svc.GetTeamByID(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		team1 := svc.createTeam(t, ctx, org)
		team2 := svc.createTeam(t, ctx, org)
		team3 := svc.createTeam(t, ctx, org)

		got, err := svc.ListTeams(ctx, org.Name)
		require.NoError(t, err)

		assert.Contains(t, got, team1)
		assert.Contains(t, got, team2)
		assert.Contains(t, got, team3)
	})

	t.Run("list members", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		org := svc.createOrganization(t, ctx)

		team := svc.createTeam(t, ctx, org)
		otherteam := svc.createTeam(t, ctx, org)
		user1 := svc.createUser(t)
		user2 := svc.createUser(t, auth.WithTeams(team))
		user3 := svc.createUser(t, auth.WithTeams(team, otherteam))

		got, err := svc.ListTeamMembers(ctx, team.ID)
		require.NoError(t, err)

		assert.Equal(t, 2, len(got), got)
		assert.NotContains(t, got, user1)
		assert.Contains(t, got, user2)
		assert.Contains(t, got, user3)
	})

	t.Run("delete", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		team := svc.createTeam(t, ctx, nil)

		err := svc.DeleteTeam(ctx, team.ID)
		require.NoError(t, err)
	})

	t.Run("disallow deleting owners team", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		org := svc.createOrganization(t, ctx) // creates owners team

		owners, err := svc.GetTeam(ctx, org.Name, "owners")
		require.NoError(t, err)

		err = svc.DeleteTeam(ctx, owners.ID)
		assert.Equal(t, auth.ErrRemovingOwnersTeamNotPermitted, err)
	})
}
