package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/daemon"
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
		svc, _, _ := setup(t, &config{Config: daemon.Config{
			Database:   connstr,
			SiteAdmins: []string{"bob", "alice", "sue"},
		}})

		areSiteAdmins := func(want bool) {
			for _, username := range []string{"bob", "alice", "sue"} {
				admin, err := svc.Users.GetUser(adminCtx, otfuser.UserSpec{Username: internal.String(username)})
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

		user := svc.createUser(t, otfuser.WithTeams(team1, team2))

		token1, _ := svc.createToken(t, ctx, user)
		_, _ = svc.createToken(t, ctx, user)

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
				spec: otfuser.UserSpec{Username: internal.String(user.Username)},
			},
			{
				name: "auth token",
				spec: otfuser.UserSpec{AuthenticationTokenID: &token1.ID},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// only admin can retrieve user info
				got, err := svc.Users.GetUser(adminCtx, tt.spec)
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
		_, err := svc.Users.GetUser(adminCtx, otfuser.UserSpec{Username: internal.String("does-not-exist")})
		assert.ErrorIs(t, err, internal.ErrResourceNotFound)
	})

	t.Run("list users", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		user1 := userFromContext(t, ctx)
		user2 := svc.createUser(t)
		user3 := svc.createUser(t)
		// only admin can retrieve its own user account
		admin := svc.getUser(t, adminCtx, otfuser.SiteAdminUsername)

		got, err := svc.Users.List(adminCtx)
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
		dev := svc.createUser(t, otfuser.WithTeams(owners, developers))

		// create guest user, member of no team
		guest := svc.createUser(t)

		got, err := svc.Users.ListOrganizationUsers(ctx, org.Name)
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
		err := svc.Users.Delete(adminCtx, user.Username)
		require.NoError(t, err)

		_, err = svc.Users.GetUser(adminCtx, otfuser.UserSpec{Username: internal.String(user.Username)})
		assert.ErrorIs(t, err, internal.ErrResourceNotFound)
	})

	t.Run("add team membership", func(t *testing.T) {
		svc, org, ctx := setup(t, nil)
		team := svc.createTeam(t, ctx, org)
		user := svc.createUser(t)

		err := svc.Users.AddTeamMembership(ctx, team.ID, []string{user.Username})
		require.NoError(t, err)

		got, err := svc.Users.GetUser(adminCtx, otfuser.UserSpec{Username: internal.String(user.Username)})
		require.NoError(t, err)

		assert.Contains(t, got.Teams, team)

		// Adding membership for a user that does not exist should first create
		// the user.
		t.Run("create new user", func(t *testing.T) {
			err := svc.Users.AddTeamMembership(ctx, team.ID, []string{"new-kid"})
			require.NoError(t, err)

			got, err := svc.Users.GetUser(adminCtx, otfuser.UserSpec{Username: internal.String("new-kid")})
			require.NoError(t, err)
			assert.Contains(t, got.Teams, team)
		})
	})

	t.Run("remove team membership", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		team := svc.createTeam(t, ctx, org)
		user := svc.createUser(t, otfuser.WithTeams(team))

		err := svc.Users.RemoveTeamMembership(ctx, team.ID, []string{user.Username})
		require.NoError(t, err)

		got, err := svc.Users.GetUser(adminCtx, otfuser.UserSpec{Username: internal.String(user.Username)})
		require.NoError(t, err)

		assert.NotContains(t, got.Teams, team)
	})

	t.Run("cannot remove last owner", func(t *testing.T) {
		// automatically creates org and owners team
		svc, org, ctx := setup(t, nil)
		owner := userFromContext(t, ctx)

		owners, err := svc.Teams.Get(ctx, org.Name, "owners")
		require.NoError(t, err)
		// add another owner
		another := svc.createUser(t, otfuser.WithTeams(owners))

		// try to delete both members from the owners team
		err = svc.Users.RemoveTeamMembership(ctx, owners.ID, []string{owner.Username, another.Username})
		assert.Equal(t, otfuser.ErrCannotDeleteOnlyOwner, err)
	})
}
