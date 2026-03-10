package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Organization(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		daemon, _, ctx := setup(t, skipDefaultOrganization())
		org, err := daemon.Organizations.CreateOrganization(ctx, organization.CreateOptions{
			Name: new(uuid.NewString()),
		})
		require.NoError(t, err)

		t.Run("duplicate error", func(t *testing.T) {
			_, err := daemon.Organizations.CreateOrganization(ctx, organization.CreateOptions{
				Name: new(org.Name.String()),
			})
			require.Equal(t, internal.ErrResourceAlreadyExists, err)
		})

		t.Run("owners team should be created", func(t *testing.T) {
			owners, err := daemon.Teams.GetTeam(ctx, org.Name, "owners")
			require.NoError(t, err)

			t.Run("creator should be a member", func(t *testing.T) {
				members, err := daemon.Users.ListTeamUsers(ctx, owners.ID)
				require.NoError(t, err)
				if assert.Equal(t, 1, len(members)) {
					user := userFromContext(t, ctx)
					assert.Equal(t, user.Username, members[0].Username)
				}
			})
		})
	})

	t.Run("update name", func(t *testing.T) {
		daemon, _, ctx := setup(t, skipDefaultOrganization())
		org := daemon.createOrganization(t, ctx)

		want := organization.NewTestName(t)
		updated, err := daemon.Organizations.UpdateOrganization(ctx, org.Name, organization.UpdateOptions{
			Name: new(want.String()),
		})
		require.NoError(t, err)

		assert.Equal(t, want, updated.Name)
	})

	t.Run("list with pagination", func(t *testing.T) {
		daemon, _, ctx := setup(t)
		_ = daemon.createOrganization(t, ctx)
		_ = daemon.createOrganization(t, ctx)

		t.Run("page one, two items per page", func(t *testing.T) {
			orgs, err := daemon.Organizations.ListOrganizations(ctx, organization.ListOptions{PageOptions: resource.PageOptions{PageNumber: 1, PageSize: 2}})
			require.NoError(t, err)

			assert.Equal(t, 2, len(orgs.Items))
		})

		t.Run("page one, one item per page", func(t *testing.T) {
			orgs, err := daemon.Organizations.ListOrganizations(ctx, organization.ListOptions{PageOptions: resource.PageOptions{PageNumber: 1, PageSize: 1}})
			require.NoError(t, err)

			assert.Equal(t, 1, len(orgs.Items))
		})

		t.Run("page two, one item per page", func(t *testing.T) {
			orgs, err := daemon.Organizations.ListOrganizations(ctx, organization.ListOptions{PageOptions: resource.PageOptions{PageNumber: 2, PageSize: 1}})
			require.NoError(t, err)

			assert.Equal(t, 1, len(orgs.Items))
		})
	})

	t.Run("list user's organizations", func(t *testing.T) {
		daemon, want1, ctx := setup(t)
		want2 := daemon.createOrganization(t, ctx)
		_ = daemon.createOrganization(t, adminCtx) // org not belonging to user

		got, err := daemon.Organizations.ListOrganizations(ctx, organization.ListOptions{})
		require.NoError(t, err)

		assert.Equal(t, 2, len(got.Items))
		assert.Contains(t, got.Items, want1)
		assert.Contains(t, got.Items, want2)
	})

	t.Run("new user should see zero orgs", func(t *testing.T) {
		daemon, _, ctx := setup(t)
		_ = daemon.createOrganization(t, ctx)
		_ = daemon.createOrganization(t, ctx)

		_, newUserCtx := daemon.createUserCtx(t)

		got, err := daemon.Organizations.ListOrganizations(newUserCtx, organization.ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 0, len(got.Items))
	})

	t.Run("get", func(t *testing.T) {
		daemon, _, ctx := setup(t)
		want := daemon.createOrganization(t, ctx)

		got, err := daemon.Organizations.GetOrganization(ctx, want.Name)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("delete", func(t *testing.T) {
		daemon, _, ctx := setup(t, skipDefaultOrganization())

		org := daemon.createOrganization(t, ctx)

		err := daemon.Organizations.DeleteOrganization(ctx, org.Name)
		require.NoError(t, err)

		_, err = daemon.Organizations.GetOrganization(ctx, org.Name)
		require.ErrorIs(t, err, internal.ErrResourceNotFound)
	})

	t.Run("delete non-existent org", func(t *testing.T) {
		daemon, _, _ := setup(t)

		// delete using site admin otherwise a not authorized error is returned
		// to normal users
		err := daemon.Organizations.DeleteOrganization(adminCtx, organization.NewTestName(t))
		require.ErrorIs(t, err, internal.ErrResourceNotFound)
	})
}
