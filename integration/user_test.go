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

func TestUser(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	svc := testutil.NewAuthService(t, db)

	t.Run("create", func(t *testing.T) {
		username := uuid.NewString()

		user, err := svc.CreateUser(ctx, username)
		require.NoError(t, err)
		svc.DeleteUser(ctx, user.ID)
	})

	t.Run("get", func(t *testing.T) {
		org1 := testutil.CreateOrganization(t, db)
		org2 := testutil.CreateOrganization(t, db)
		team1 := testutil.CreateTeam(t, db, org1)
		team2 := testutil.CreateTeam(t, db, org2)

		memberships := []auth.NewUserOption{
			auth.WithOrganizations(org1.Name(), org2.Name()),
			auth.WithTeams(team1, team2),
		}
		user := testutil.CreateUser(t, db, memberships...)
		session := testutil.CreateSession(t, db, user.ID, nil)

		tests := []struct {
			name string
			spec otf.UserSpec
		}{
			{
				name: "id",
				spec: otf.UserSpec{UserID: otf.String(user.ID)},
			},
			{
				name: "username",
				spec: otf.UserSpec{Username: otf.String(user.Username())},
			},
			{
				name: "session token",
				spec: otf.UserSpec{SessionToken: otf.String(session.Token())},
			},
			{
				name: "auth token",
				spec: otf.UserSpec{AuthenticationToken: otf.String(token1.Token())},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := svc.GetUser(ctx, tt.spec)
				require.NoError(t, err)

				assert.Equal(t, got.ID, user.ID)
				assert.Equal(t, got.Username(), user.Username())
				assert.Equal(t, 2, len(got.Organizations()))
				assert.Equal(t, 2, len(got.Teams()))
			})
		}
	})

	t.Run("add organization membership", func(t *testing.T) {
		org := testutil.CreateOrganization(t, db)
		user := testutil.CreateUser(t, db)

		err := svc.AddOrganizationMembership(ctx, user.ID, org.Name())
		require.NoError(t, err)

		got, err := svc.GetUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
		require.NoError(t, err)

		assert.Contains(t, got.Organizations(), org.Name())
	})

	t.Run("remove organization membership", func(t *testing.T) {
		org := testutil.CreateOrganization(t, db)
		user := testutil.CreateUser(t, db)

		err := svc.RemoveOrganizationMembership(ctx, user.ID, org.Name())
		require.NoError(t, err)

		got, err := svc.GetUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
		require.NoError(t, err)

		assert.NotContains(t, got.Organizations(), org)
	})

	t.Run("delete", func(t *testing.T) {
		user := testutil.CreateUser(t, db)

		err := svc.DeleteUser(ctx, user.ID)
		require.NoError(t, err)
	})
}
