package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamDB(t *testing.T) {
	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})

	t.Run("create", func(t *testing.T) {
		svc := setup(t, "")
		org := svc.createOrganization(t, ctx)

		team, err := svc.CreateTeam(ctx, auth.NewTeamOptions{
			Name:         uuid.NewString(),
			Organization: org.Name,
		})
		require.NoError(t, err)

		t.Run("already exists error", func(t *testing.T) {
			_, err := svc.CreateTeam(ctx, auth.NewTeamOptions{
				Name:         team.Name,
				Organization: org.Name,
			})
			require.Equal(t, otf.ErrResourceAlreadyExists, err)
		})
	})

	t.Run("update", func(t *testing.T) {
		svc := setup(t, "")
		team := svc.createTeam(t, ctx, nil)

		_, err := svc.UpdateTeam(ctx, team.ID, auth.UpdateTeamOptions{
			OrganizationAccess: auth.OrganizationAccess{
				ManageWorkspaces: true,
				ManageVCS:        true,
				ManageRegistry:   true,
			},
		})
		require.NoError(t, err)

		got, err := svc.GetTeam(ctx, team.Organization, team.Name)
		require.NoError(t, err)

		assert.True(t, got.OrganizationAccess().ManageWorkspaces)
		assert.True(t, got.OrganizationAccess().ManageVCS)
		assert.True(t, got.OrganizationAccess().ManageRegistry)
	})

	t.Run("get", func(t *testing.T) {
		svc := setup(t, "")
		team := svc.createTeam(t, ctx, nil)

		got, err := svc.GetTeam(ctx, team.Organization, team.Name)
		require.NoError(t, err)

		assert.Equal(t, team, got)
	})

	t.Run("get by id", func(t *testing.T) {
		svc := setup(t, "")
		want := svc.createTeam(t, ctx, nil)

		got, err := svc.GetTeamByID(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		svc := setup(t, "")
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
		svc := setup(t, "")
		org := svc.createOrganization(t, ctx)

		team := svc.createTeam(t, ctx, org)
		user1 := svc.createUser(t, ctx, auth.WithOrganizations(org.Name))
		user2 := svc.createUser(t, ctx, auth.WithOrganizations(org.Name), auth.WithTeams(team))
		user3 := svc.createUser(t, ctx, auth.WithOrganizations(org.Name), auth.WithTeams(team))

		got, err := svc.ListTeamMembers(ctx, team.ID)
		require.NoError(t, err)

		assert.NotContains(t, got, user1)
		assert.Contains(t, got, user2)
		assert.Contains(t, got, user3)
	})

	t.Run("delete", func(t *testing.T) {
		svc := setup(t, "")
		team := svc.createTeam(t, ctx, nil)

		err := svc.DeleteTeam(ctx, team.ID)
		require.NoError(t, err)
	})
}
