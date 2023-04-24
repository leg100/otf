package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/daemon"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &auth.SiteAdmin)

	t.Run("set site admins", func(t *testing.T) {
		connstr := sql.NewTestDB(t)
		svc := setup(t, &config{Config: daemon.Config{
			Database:   connstr,
			SiteAdmins: []string{"bob", "alice", "sue"},
		}})

		areSiteAdmins := func(want bool) {
			for _, username := range []string{"bob", "alice", "sue"} {
				admin, err := svc.GetUser(ctx, auth.UserSpec{Username: otf.String(username)})
				require.NoError(t, err)
				assert.Equal(t, want, admin.IsSiteAdmin())
			}
		}
		areSiteAdmins(true)

		// Start another daemon with *no* site admins specified and verify that
		// bob, alice and sue are no longer site admins.
		t.Run("reset", func(t *testing.T) {
			svc = setup(t, &config{Config: daemon.Config{
				Database: connstr,
			}})
			areSiteAdmins(false)
		})
	})

	t.Run("get", func(t *testing.T) {
		svc := setup(t, nil)

		org1 := svc.createOrganization(t, ctx)
		org2 := svc.createOrganization(t, ctx)
		team1 := svc.createTeam(t, ctx, org1)
		team2 := svc.createTeam(t, ctx, org2)

		user := svc.createUser(t, ctx,
			auth.WithTeams(team1, team2))

		token1, _ := svc.createToken(t, ctx, user)
		_, _ = svc.createToken(t, ctx, user)

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
				name: "auth token",
				spec: auth.UserSpec{AuthenticationTokenID: otf.String(token1.ID)},
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
				assert.Equal(t, 2, len(got.Organizations()))
				assert.Equal(t, 2, len(got.Teams))
			})
		}
	})

	t.Run("get not found error", func(t *testing.T) {
		svc := setup(t, nil)
		_, err := svc.GetUser(ctx, auth.UserSpec{Username: otf.String("does-not-exist")})
		assert.Equal(t, otf.ErrResourceNotFound, err)
	})

	// List users in an organization. The underlying SQL joins users to
	// organization via teams, so this test adds a user to one team and another
	// user to two teams, with both teams in the same organization, to check the
	// SQL is working correctly, e.g. performing not only the join correctly,
	// but performing de-duplication too so that users are not listed more than
	// once.
	t.Run("list", func(t *testing.T) {
		svc := setup(t, nil)
		// create owners team consisting of one owner
		owner, userCtx := svc.createUserCtx(t, ctx)
		org := svc.createOrganization(t, userCtx)
		owner = svc.getUser(t, ctx, owner.Username) // refresh user to update team membership
		owners := svc.getTeam(t, ctx, org.Name, "owners")

		// create developers team
		developers := svc.createTeam(t, ctx, org)

		// add dev user to both teams
		dev := svc.createUser(t, ctx, auth.WithTeams(owners, developers))

		// create guest user, member of no team
		guest := svc.createUser(t, ctx)

		got, err := svc.ListOrganizationUsers(ctx, org.Name)
		require.NoError(t, err)

		// should get list of two users: owner and dev
		assert.Equal(t, 2, len(got), got)
		assert.Contains(t, got, owner)
		assert.Contains(t, got, dev)
		assert.NotContains(t, got, guest)
	})

	t.Run("delete", func(t *testing.T) {
		svc := setup(t, nil)
		user := svc.createUser(t, ctx)

		err := svc.DeleteUser(ctx, user.Username)
		require.NoError(t, err)

		_, err = svc.GetUser(ctx, auth.UserSpec{Username: otf.String(user.Username)})
		assert.Equal(t, err, otf.ErrResourceNotFound)
	})

	t.Run("add team membership", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		team := svc.createTeam(t, ctx, org)
		user := svc.createUser(t, ctx)

		err := svc.AddTeamMembership(ctx, auth.TeamMembershipOptions{
			Username: user.Username,
			TeamID:   team.ID,
		})
		require.NoError(t, err)

		got, err := svc.GetUser(ctx, auth.UserSpec{Username: otf.String(user.Username)})
		require.NoError(t, err)

		assert.Contains(t, got.Teams, team)
	})

	t.Run("remove team membership", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		team := svc.createTeam(t, ctx, org)
		user := svc.createUser(t, ctx, auth.WithTeams(team))

		err := svc.RemoveTeamMembership(ctx, auth.TeamMembershipOptions{
			Username: user.Username,
			TeamID:   team.ID,
		})
		require.NoError(t, err)

		got, err := svc.GetUser(ctx, auth.UserSpec{Username: otf.String(user.Username)})
		require.NoError(t, err)

		assert.NotContains(t, got.Teams, team)
	})

	t.Run("cannot remove last owner", func(t *testing.T) {
		svc := setup(t, nil)
		// automatically creates owners team with site admin as owner
		org := svc.createOrganization(t, ctx)

		owners, err := svc.GetTeam(ctx, org.Name, "owners")
		require.NoError(t, err)

		err = svc.RemoveTeamMembership(ctx, auth.TeamMembershipOptions{
			Username: auth.SiteAdminUsername,
			TeamID:   owners.ID,
		})
		assert.Equal(t, auth.ErrCannotDeleteOnlyOwner, err)
	})
}
