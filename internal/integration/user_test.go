package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
	t.Parallel()

	// Tests the --site-admins functionality, promoting a number of users to the
	// role of site admin when starting the daemon.
	t.Run("set site admins", func(t *testing.T) {
		connstr := sql.NewTestDB(t)
		svc, _, _ := setup(t, &config{Config: daemon.Config{
			Database:   connstr,
			SiteAdmins: []string{"bob", "alice", "sue"},
		}})

		areSiteAdmins := func(want bool) {
			for _, username := range []string{"bob", "alice", "sue"} {
				admin, err := svc.GetUser(adminCtx, auth.UserSpec{Username: internal.String(username)})
				require.NoError(t, err)
				assert.Equal(t, want, admin.IsSiteAdmin())
			}
		}
		areSiteAdmins(true)

		// Start another daemon with *no* site admins specified, which should
		// relegate the users back to normal users.
		t.Run("reset", func(t *testing.T) {
			svc, _, _ = setup(t, &config{Config: daemon.Config{
				Database: connstr,
			}})
			areSiteAdmins(false)
		})
	})

	// Create a user and a user token and test retrieving the user using their ID, username and
	// token.
	t.Run("get", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)

		org1 := svc.createOrganization(t, ctx)
		org2 := svc.createOrganization(t, ctx)
		team1 := svc.createTeam(t, ctx, org1)
		team2 := svc.createTeam(t, ctx, org2)

		user := svc.createUser(t, auth.WithTeams(team1, team2))

		token1, _ := svc.createToken(t, ctx, user)
		_, _ = svc.createToken(t, ctx, user)

		tests := []struct {
			name string
			spec auth.UserSpec
		}{
			{
				name: "id",
				spec: auth.UserSpec{UserID: internal.String(user.ID)},
			},
			{
				name: "username",
				spec: auth.UserSpec{Username: internal.String(user.Username)},
			},
			{
				name: "auth token",
				spec: auth.UserSpec{AuthenticationTokenID: internal.String(token1.ID)},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// only admin can retrieve user info
				got, err := svc.GetUser(adminCtx, tt.spec)
				require.NoError(t, err)

				assert.Equal(t, got.ID, user.ID)
				assert.Equal(t, got.Username, user.Username)
				assert.Equal(t, got.CreatedAt, user.CreatedAt)
				assert.Equal(t, got.UpdatedAt, user.UpdatedAt)
				assert.Equal(t, 2, len(got.Organizations()))
				assert.Equal(t, 2, len(got.Teams))
			})
		}
	})

	t.Run("get not found error", func(t *testing.T) {
		svc, _, _ := setup(t, nil)
		_, err := svc.GetUser(adminCtx, auth.UserSpec{Username: internal.String("does-not-exist")})
		assert.Equal(t, internal.ErrResourceNotFound, err)
	})

	t.Run("list users", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		user1 := userFromContext(t, ctx)
		user2 := svc.createUser(t)
		user3 := svc.createUser(t)
		// only admin can retrieve its own user account
		admin := svc.getUser(t, adminCtx, auth.SiteAdminUsername)

		got, err := svc.ListUsers(ctx)
		require.NoError(t, err)

		assert.Equal(t, 4, len(got))
		assert.Contains(t, got, user1)
		assert.Contains(t, got, user2)
		assert.Contains(t, got, user3)
		assert.Contains(t, got, admin)
	})

	// List users in an organization. The underlying SQL joins users to
	// organization via teams, so this test adds a user to one team and another
	// user to two teams, with both teams in the same organization, to check the
	// SQL is working correctly, e.g. performing not only the join correctly,
	// but performing de-duplication too so that users are not listed more than
	// once.
	t.Run("list organization users", func(t *testing.T) {
		// automatically creates owners team consisting of one owner
		svc, org, ctx := setup(t, nil)
		owner := userFromContext(t, ctx)
		owners := svc.getTeam(t, ctx, org.Name, "owners")

		// create developers team
		developers := svc.createTeam(t, ctx, org)

		// create dev user and add to both teams
		dev := svc.createUser(t, auth.WithTeams(owners, developers))

		// create guest user, member of no team
		guest := svc.createUser(t)

		got, err := svc.ListOrganizationUsers(ctx, org.Name)
		require.NoError(t, err)

		// should get list of two users: owner and dev
		assert.Equal(t, 2, len(got), got)
		assert.Contains(t, got, owner)
		assert.Contains(t, got, dev)
		assert.NotContains(t, got, guest)
	})

	t.Run("delete", func(t *testing.T) {
		svc, _, _ := setup(t, nil)
		user := svc.createUser(t)

		// only admin can delete user
		err := svc.DeleteUser(adminCtx, user.Username)
		require.NoError(t, err)

		_, err = svc.GetUser(adminCtx, auth.UserSpec{Username: internal.String(user.Username)})
		assert.Equal(t, err, internal.ErrResourceNotFound)
	})

	t.Run("add team membership", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		team := svc.createTeam(t, ctx, org)
		user := svc.createUser(t)

		err := svc.AddTeamMembership(ctx, auth.TeamMembershipOptions{
			Username: user.Username,
			TeamID:   team.ID,
		})
		require.NoError(t, err)

		got, err := svc.GetUser(adminCtx, auth.UserSpec{Username: internal.String(user.Username)})
		require.NoError(t, err)

		assert.Contains(t, got.Teams, team)
	})

	t.Run("remove team membership", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		team := svc.createTeam(t, ctx, org)
		user := svc.createUser(t, auth.WithTeams(team))

		err := svc.RemoveTeamMembership(ctx, auth.TeamMembershipOptions{
			Username: user.Username,
			TeamID:   team.ID,
		})
		require.NoError(t, err)

		got, err := svc.GetUser(adminCtx, auth.UserSpec{Username: internal.String(user.Username)})
		require.NoError(t, err)

		assert.NotContains(t, got.Teams, team)
	})

	t.Run("cannot remove last owner", func(t *testing.T) {
		// automatically creates org and owners team
		svc, org, ctx := setup(t, nil)
		owner := userFromContext(t, ctx)

		owners, err := svc.GetTeam(ctx, org.Name, "owners")
		require.NoError(t, err)

		err = svc.RemoveTeamMembership(ctx, auth.TeamMembershipOptions{
			Username: owner.Username,
			TeamID:   owners.ID,
		})
		assert.Equal(t, auth.ErrCannotDeleteOnlyOwner, err)
	})
}
