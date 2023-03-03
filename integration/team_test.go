package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeam(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	org := testutil.CreateOrganization(t, db)
	svc := testutil.NewAuthService(t, db)

	t.Run("create", func(t *testing.T) {
		team, err := svc.CreateTeam(ctx, otf.CreateTeamOptions{
			Name:         uuid.NewString(),
			Organization: org.Name(),
		})
		require.NoError(t, err)
		defer svc.DeleteTeam(ctx, team.ID)

		t.Run("duplicate name", func(t *testing.T) {
			_, err := svc.CreateTeam(ctx, otf.CreateTeamOptions{
				Name:         uuid.NewString(),
				Organization: org.Name(),
			})
			require.Equal(t, otf.ErrResourceAlreadyExists, err)
		})
	})
	t.Run("list", func(t *testing.T) {
		team1 := testutil.CreateTeam(t, db, org)
		team2 := testutil.CreateTeam(t, db, org)
		team3 := testutil.CreateTeam(t, db, org)

		got, err := svc.ListTeams(ctx, org.Name())
		require.NoError(t, err)

		assert.Contains(t, got, team1)
		assert.Contains(t, got, team2)
		assert.Contains(t, got, team3)
	})

	t.Run("list members", func(t *testing.T) {
		team := testutil.CreateTeam(t, db, org)

		memberships := []auth.NewUserOption{
			auth.WithOrganizations(org.Name()),
			auth.WithTeams(team),
		}
		user1 := testutil.CreateUser(t, db, memberships...)
		user2 := testutil.CreateUser(t, db, memberships...)

		got, err := svc.ListTeamMembers(context.Background(), team.ID)
		require.NoError(t, err)

		assert.Contains(t, got, user1)
		assert.Contains(t, got, user2)
	})

	t.Run("update", func(t *testing.T) {
		team := testutil.CreateTeam(t, db, org)

		got, err := svc.UpdateTeam(ctx, team.ID, auth.UpdateTeamOptions{
			OrganizationAccess: auth.OrganizationAccess{
				ManageWorkspaces: true,
				ManageVCS:        true,
				ManageRegistry:   true,
			},
		})
		require.NoError(t, err)

		assert.True(t, got.OrganizationAccess().ManageWorkspaces)
		assert.True(t, got.OrganizationAccess().ManageVCS)
		assert.True(t, got.OrganizationAccess().ManageRegistry)
	})

	t.Run("get by name", func(t *testing.T) {
		team := testutil.CreateTeam(t, db, org)

		got, err := svc.GetTeam(ctx, team.Name(), org.Name())
		require.NoError(t, err)

		assert.Equal(t, team, got)
	})

	t.Run("get by id", func(t *testing.T) {
		want := testutil.CreateTeam(t, db, org)

		got, err := svc.GetTeamByID(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
}
