package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	otfuser "github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
	integrationTest(t)

	// Tests the --site-admins functionality, promoting a number of users to the
	// role of site admin when starting the daemon.
	t.Run("set site admins", func(t *testing.T) {
		connstr := sql.NewTestDB(t)
		daemon, _, _ := setup(t, withDatabase(connstr), withSiteAdmins("bob", "alice", "sue"))

		areSiteAdmins := func(want bool) {
			for _, usernameStr := range []string{"bob", "alice", "sue"} {
				username := otfuser.MustUsername(usernameStr)
				admin, err := daemon.Users.GetUser(adminCtx, otfuser.UserSpec{Username: &username})
				require.NoError(t, err)
				assert.Equal(t, want, admin.IsSiteAdmin())
			}
		}
		areSiteAdmins(true)

		// Start another daemon with *no* site admins specified, which should
		// relegate the users back to normal users.
		t.Run("reset", func(t *testing.T) {
			daemon, _, _ = setup(t, withDatabase(connstr))
			areSiteAdmins(false)
		})
	})

	// Create a user and a user token and test retrieving the user using their ID, username and
	// token.
	t.Run("get", func(t *testing.T) {
		daemon, _, ctx := setup(t)

		org1 := daemon.createOrganization(t, ctx)
		org2 := daemon.createOrganization(t, ctx)
		team1 := daemon.createTeam(t, ctx, org1)
		team2 := daemon.createTeam(t, ctx, org2)

		user := daemon.createUser(t, otfuser.WithTeams(team1, team2))

		token1, _ := daemon.createToken(t, ctx, user)
		_, _ = daemon.createToken(t, ctx, user)

		tests := []struct {
			name string
			spec otfuser.UserSpec
		}{
			{
				name: "id",
				spec: otfuser.UserSpec{UserID: &user.ID},
			},
			{
				name: "username",
				spec: otfuser.UserSpec{Username: &user.Username},
			},
			{
				name: "auth token",
				spec: otfuser.UserSpec{AuthenticationTokenID: &token1.ID},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// only admin can retrieve user info
				got, err := daemon.Users.GetUser(adminCtx, tt.spec)
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
		daemon, _, _ := setup(t)
		unknown := otfuser.NewTestUsername(t)
		_, err := daemon.Users.GetUser(adminCtx, otfuser.UserSpec{Username: &unknown})
		assert.ErrorIs(t, err, internal.ErrResourceNotFound)
	})

	t.Run("list users", func(t *testing.T) {
		daemon, _, ctx := setup(t)
		user1 := userFromContext(t, ctx)
		user2 := daemon.createUser(t)
		user3 := daemon.createUser(t)
		// only admin can retrieve its own user account
		admin := daemon.getUser(t, adminCtx, otfuser.SiteAdminUsername)

		got, err := daemon.Users.List(adminCtx)
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
		daemon, org, ctx := setup(t)
		owner := userFromContext(t, ctx)
		owners := daemon.getTeam(t, ctx, org.Name, "owners")

		// create developers team
		developers := daemon.createTeam(t, ctx, org)

		// create dev user and add to both teams
		dev := daemon.createUser(t, otfuser.WithTeams(owners, developers))

		// create guest user, member of no team
		guest := daemon.createUser(t)

		got, err := daemon.Users.ListOrganizationUsers(ctx, org.Name)
		require.NoError(t, err)

		// should get list of two users: owner and dev
		assert.Equal(t, 2, len(got), got)
		assert.Contains(t, got, owner)
		assert.Contains(t, got, dev)
		assert.NotContains(t, got, guest)
	})

	t.Run("delete", func(t *testing.T) {
		daemon, _, _ := setup(t)
		user := daemon.createUser(t)

		// only admin can delete user
		err := daemon.Users.Delete(adminCtx, user.Username)
		require.NoError(t, err)

		_, err = daemon.Users.GetUser(adminCtx, otfuser.UserSpec{Username: &user.Username})
		assert.ErrorIs(t, err, internal.ErrResourceNotFound)
	})

	t.Run("add team membership", func(t *testing.T) {
		daemon, org, ctx := setup(t)
		team := daemon.createTeam(t, ctx, org)
		user := daemon.createUser(t)

		err := daemon.Users.AddTeamMembership(ctx, team.ID, []otfuser.Username{user.Username})
		require.NoError(t, err)

		got, err := daemon.Users.GetUser(adminCtx, otfuser.UserSpec{Username: &user.Username})
		require.NoError(t, err)

		assert.Contains(t, got.Teams, team)

		// Adding membership for a user that does not exist should first create
		// the user.
		t.Run("create new user", func(t *testing.T) {
			username := otfuser.MustUsername("new-kid")
			err := daemon.Users.AddTeamMembership(ctx, team.ID, []otfuser.Username{username})
			require.NoError(t, err)

			got, err := daemon.Users.GetUser(adminCtx, otfuser.UserSpec{Username: &username})
			require.NoError(t, err)
			assert.Contains(t, got.Teams, team)
		})
	})

	t.Run("remove team membership", func(t *testing.T) {
		daemon, _, ctx := setup(t)
		org := daemon.createOrganization(t, ctx)
		team := daemon.createTeam(t, ctx, org)
		user := daemon.createUser(t, otfuser.WithTeams(team))

		err := daemon.Users.RemoveTeamMembership(ctx, team.ID, []otfuser.Username{user.Username})
		require.NoError(t, err)

		got, err := daemon.Users.GetUser(adminCtx, otfuser.UserSpec{Username: &user.Username})
		require.NoError(t, err)

		assert.NotContains(t, got.Teams, team)
	})

	t.Run("cannot remove last owner", func(t *testing.T) {
		// automatically creates org and owners team
		daemon, org, ctx := setup(t)
		owner := userFromContext(t, ctx)

		owners, err := daemon.Teams.Get(ctx, org.Name, "owners")
		require.NoError(t, err)
		// add another owner
		another := daemon.createUser(t, otfuser.WithTeams(owners))

		// try to delete both members from the owners team
		err = daemon.Users.RemoveTeamMembership(ctx, owners.ID, []otfuser.Username{owner.Username, another.Username})
		assert.Equal(t, otfuser.ErrCannotDeleteOnlyOwner, err)
	})
}
