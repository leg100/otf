package auth

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamDB(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	t.Run("create", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		team := NewTeam(NewTeamOptions{Name: "team-awesome", Organization: org.Name})

		defer db.deleteTeam(ctx, team.ID)

		err := db.createTeam(ctx, team)
		require.NoError(t, err)

		t.Run("duplicate name", func(t *testing.T) {
			dup := NewTeam(NewTeamOptions{Name: "team-awesome", Organization: org.Name})
			err := db.createTeam(ctx, dup)
			require.Equal(t, otf.ErrResourceAlreadyExists, err)
		})
	})

	t.Run("update", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		team := createTestTeam(t, db, org.Name)

		_, err := db.UpdateTeam(ctx, team.ID, func(team *Team) error {
			return team.Update(UpdateTeamOptions{
				OrganizationAccess: OrganizationAccess{
					ManageWorkspaces: true,
					ManageVCS:        true,
					ManageRegistry:   true,
				},
			})
		})
		require.NoError(t, err)

		got, err := db.getTeam(ctx, team.Name, org.Name)
		require.NoError(t, err)

		assert.True(t, got.OrganizationAccess().ManageWorkspaces)
		assert.True(t, got.OrganizationAccess().ManageVCS)
		assert.True(t, got.OrganizationAccess().ManageRegistry)
	})

	t.Run("get", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		team := createTestTeam(t, db, org.Name)

		got, err := db.getTeam(ctx, team.Name, org.Name)
		require.NoError(t, err)

		assert.Equal(t, team, got)
	})

	t.Run("get by id", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		want := createTestTeam(t, db, org.Name)

		got, err := db.getTeamByID(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		team1 := createTestTeam(t, db, org.Name)
		team2 := createTestTeam(t, db, org.Name)
		team3 := createTestTeam(t, db, org.Name)

		got, err := db.listTeams(ctx, org.Name)
		require.NoError(t, err)

		assert.Contains(t, got, team1)
		assert.Contains(t, got, team2)
		assert.Contains(t, got, team3)
	})

	t.Run("list members", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		team := createTestTeam(t, db, org.Name)
		user1 := createTestUser(t, db, WithOrganizations(org.Name))
		user2 := createTestUser(t, db, WithOrganizations(org.Name), WithTeams(team))
		user3 := createTestUser(t, db, WithOrganizations(org.Name), WithTeams(team))

		got, err := db.listTeamMembers(ctx, team.ID)
		require.NoError(t, err)

		assert.NotContains(t, got, user1)
		assert.Contains(t, got, user2)
		assert.Contains(t, got, user3)
	})

	t.Run("delete", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		team := createTestTeam(t, db, org.Name)

		err := db.deleteTeam(ctx, team.ID)
		require.NoError(t, err)
	})
}
