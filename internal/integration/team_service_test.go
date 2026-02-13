package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	otfteam "github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegation_TeamService(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		svc, org, ctx := setup(t)

		team, err := svc.Teams.Create(ctx, org.Name, otfteam.CreateTeamOptions{
			Name: new(uuid.NewString()),
		})
		require.NoError(t, err)

		t.Run("already exists error", func(t *testing.T) {
			_, err := svc.Teams.Create(ctx, org.Name, otfteam.CreateTeamOptions{
				Name: new(team.Name),
			})
			require.Equal(t, internal.ErrResourceAlreadyExists, err)
		})
	})

	t.Run("update", func(t *testing.T) {
		svc, _, ctx := setup(t)
		team := svc.createTeam(t, ctx, nil)

		_, err := svc.Teams.Update(ctx, team.ID, otfteam.UpdateTeamOptions{
			OrganizationAccessOptions: otfteam.OrganizationAccessOptions{
				ManageWorkspaces: new(true),
				ManageVCS:        new(true),
				ManageModules:    new(true),
			},
		})
		require.NoError(t, err)

		got, err := svc.Teams.Get(ctx, team.Organization, team.Name)
		require.NoError(t, err)

		assert.True(t, got.ManageWorkspaces)
		assert.True(t, got.ManageVCS)
		assert.True(t, got.ManageModules)
	})

	t.Run("get", func(t *testing.T) {
		svc, _, ctx := setup(t)
		team := svc.createTeam(t, ctx, nil)

		got, err := svc.Teams.Get(ctx, team.Organization, team.Name)
		require.NoError(t, err)

		assert.Equal(t, team, got)
	})

	t.Run("get by id", func(t *testing.T) {
		svc, _, ctx := setup(t)
		want := svc.createTeam(t, ctx, nil)

		got, err := svc.Teams.GetByID(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		svc, _, ctx := setup(t)
		org := svc.createOrganization(t, ctx)
		team1 := svc.createTeam(t, ctx, org)
		team2 := svc.createTeam(t, ctx, org)
		team3 := svc.createTeam(t, ctx, org)

		got, err := svc.Teams.List(ctx, org.Name)
		require.NoError(t, err)

		assert.Contains(t, got, team1)
		assert.Contains(t, got, team2)
		assert.Contains(t, got, team3)
	})

	t.Run("list members", func(t *testing.T) {
		svc, _, ctx := setup(t)
		org := svc.createOrganization(t, ctx)

		team := svc.createTeam(t, ctx, org)
		otherteam := svc.createTeam(t, ctx, org)
		user1 := svc.createUser(t)
		user2 := svc.createUser(t, user.WithTeams(team))
		user3 := svc.createUser(t, user.WithTeams(team, otherteam))

		got, err := svc.Users.ListTeamUsers(ctx, team.ID)
		require.NoError(t, err)

		assert.Equal(t, 2, len(got), got)
		assert.NotContains(t, got, user1)
		assert.Contains(t, got, user2)
		assert.Contains(t, got, user3)
	})

	t.Run("delete", func(t *testing.T) {
		svc, _, ctx := setup(t)
		team := svc.createTeam(t, ctx, nil)

		err := svc.Teams.Delete(ctx, team.ID)
		require.NoError(t, err)
	})

	t.Run("disallow deleting owners team", func(t *testing.T) {
		svc, _, ctx := setup(t)
		org := svc.createOrganization(t, ctx) // creates owners team

		owners, err := svc.Teams.Get(ctx, org.Name, "owners")
		require.NoError(t, err)

		err = svc.Teams.Delete(ctx, owners.ID)
		assert.Equal(t, otfteam.ErrRemovingOwnersTeamNotPermitted, err)
	})
}
