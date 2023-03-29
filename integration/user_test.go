package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})

	t.Run("get", func(t *testing.T) {
		svc := setup(t, nil)

		org1 := svc.createOrganization(t, ctx)
		org2 := svc.createOrganization(t, ctx)
		team1 := svc.createTeam(t, ctx, org1)
		team2 := svc.createTeam(t, ctx, org2)

		user := svc.createUser(t, ctx,
			auth.WithOrganizations(org1.Name, org2.Name),
			auth.WithTeams(team1, team2))

		session1 := svc.createSession(t, ctx, user, nil)
		_ = svc.createSession(t, ctx, user, nil)

		token1 := svc.createToken(t, ctx, user)
		_ = svc.createToken(t, ctx, user)

		tests := []struct {
			name string
			spec auth.UserSpec
		}{
			{
				name: "id",
				spec: auth.UserSpec{UserID: otf.String(user.ID)},
			},
			{
				name: "username",
				spec: auth.UserSpec{Username: otf.String(user.Username)},
			},
			{
				name: "session token",
				spec: auth.UserSpec{SessionToken: otf.String(session1.Token())},
			},
			{
				name: "auth token",
				spec: auth.UserSpec{AuthenticationToken: otf.String(token1.Token)},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := svc.GetUser(ctx, tt.spec)
				require.NoError(t, err)

				assert.Equal(t, got.ID, user.ID)
				assert.Equal(t, got.Username, user.Username)
				assert.Equal(t, got.CreatedAt, user.CreatedAt)
				assert.Equal(t, got.UpdatedAt, user.UpdatedAt)
				assert.Equal(t, 2, len(got.Organizations))
				assert.Equal(t, 2, len(got.Teams))
			})
		}
	})

	t.Run("get not found error", func(t *testing.T) {
		svc := setup(t, nil)
		_, err := svc.GetUser(ctx, auth.UserSpec{Username: otf.String("does-not-exist")})
		assert.Equal(t, otf.ErrResourceNotFound, err)
	})

	t.Run("list", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		user1 := svc.createUser(t, ctx)
		user2 := svc.createUser(t, ctx, auth.WithOrganizations(org.Name))
		user3 := svc.createUser(t, ctx, auth.WithOrganizations(org.Name))

		users, err := svc.ListUsers(ctx, org.Name)
		require.NoError(t, err)

		assert.NotContains(t, users, user1)
		assert.Contains(t, users, user2)
		assert.Contains(t, users, user3)
	})

	t.Run("delete", func(t *testing.T) {
		svc := setup(t, nil)
		user := svc.createUser(t, ctx)

		err := svc.DeleteUser(ctx, user.Username)
		require.NoError(t, err)

		_, err = svc.GetUser(ctx, auth.UserSpec{Username: otf.String(user.Username)})
		assert.Equal(t, err, otf.ErrResourceNotFound)
	})

	t.Run("add organization membership", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		user := svc.createUser(t, ctx)

		err := svc.AddOrganizationMembership(ctx, user.ID, org.Name)
		require.NoError(t, err)

		got, err := svc.GetUser(ctx, auth.UserSpec{Username: otf.String(user.Username)})
		require.NoError(t, err)

		assert.Contains(t, got.Organizations, org.Name)
	})

	t.Run("remove organization membership", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		user := svc.createUser(t, ctx, auth.WithOrganizations(org.Name))

		err := svc.RemoveOrganizationMembership(ctx, user.ID, org.Name)
		require.NoError(t, err)

		got, err := svc.GetUser(ctx, auth.UserSpec{Username: otf.String(user.Username)})
		require.NoError(t, err)

		assert.NotContains(t, got.Organizations, org.Name)
	})

	t.Run("add team membership", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		team := svc.createTeam(t, ctx, org)
		user := svc.createUser(t, ctx, auth.WithOrganizations(org.Name))

		err := svc.AddTeamMembership(ctx, user.ID, team.ID)
		require.NoError(t, err)

		got, err := svc.GetUser(ctx, auth.UserSpec{Username: otf.String(user.Username)})
		require.NoError(t, err)

		assert.Contains(t, got.Teams, team)
	})

	t.Run("remove team membership", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		team := svc.createTeam(t, ctx, org)
		user := svc.createUser(t, ctx, auth.WithOrganizations(org.Name), auth.WithTeams(team))

		err := svc.RemoveTeamMembership(ctx, user.ID, team.ID)
		require.NoError(t, err)

		got, err := svc.GetUser(ctx, auth.UserSpec{Username: otf.String(user.Username)})
		require.NoError(t, err)

		assert.NotContains(t, got.Teams, team)
	})
}
