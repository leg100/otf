package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/orgcreator"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := internal.AddSubjectToContext(context.Background(), &auth.SiteAdmin)

	t.Run("create", func(t *testing.T) {
		svc := setup(t, nil)
		sub := svc.createSubscriber(t, ctx)
		org, err := svc.CreateOrganization(ctx, orgcreator.OrganizationCreateOptions{
			Name: internal.String(uuid.NewString()),
		})
		require.NoError(t, err)

		t.Run("duplicate error", func(t *testing.T) {
			_, err := svc.CreateOrganization(ctx, orgcreator.OrganizationCreateOptions{
				Name: internal.String(org.Name),
			})
			require.Equal(t, internal.ErrResourceAlreadyExists, err)
		})

		t.Run("owners team should be created", func(t *testing.T) {
			owners, err := svc.GetTeam(ctx, org.Name, "owners")
			require.NoError(t, err)

			t.Run("creator should be a member", func(t *testing.T) {
				members, err := svc.ListTeamMembers(ctx, owners.ID)
				require.NoError(t, err)
				if assert.Equal(t, 1, len(members)) {
					assert.Equal(t, auth.SiteAdminUsername, members[0].Username)
				}
			})
		})

		t.Run("receive event", func(t *testing.T) {
			assert.Equal(t, pubsub.NewCreatedEvent(org), <-sub)
		})
	})

	t.Run("update name", func(t *testing.T) {
		svc := setup(t, nil)
		sub := svc.createSubscriber(t, ctx)
		org := svc.createOrganization(t, ctx)
		assert.Equal(t, pubsub.NewCreatedEvent(org), <-sub)

		want := uuid.NewString()
		updated, err := svc.UpdateOrganization(ctx, org.Name, organization.OrganizationUpdateOptions{
			Name: internal.String(want),
		})
		require.NoError(t, err)

		assert.Equal(t, want, updated.Name)
		assert.Equal(t, pubsub.NewUpdatedEvent(updated), <-sub)
	})

	t.Run("list with pagination", func(t *testing.T) {
		svc := setup(t, nil)
		_ = svc.createOrganization(t, ctx)
		_ = svc.createOrganization(t, ctx)

		t.Run("page one, two items per page", func(t *testing.T) {
			orgs, err := svc.ListOrganizations(ctx, organization.OrganizationListOptions{ListOptions: internal.ListOptions{PageNumber: 1, PageSize: 2}})
			require.NoError(t, err)

			assert.Equal(t, 2, len(orgs.Items))
		})

		t.Run("page one, one item per page", func(t *testing.T) {
			orgs, err := svc.ListOrganizations(ctx, organization.OrganizationListOptions{ListOptions: internal.ListOptions{PageNumber: 1, PageSize: 1}})
			require.NoError(t, err)

			assert.Equal(t, 1, len(orgs.Items))
		})

		t.Run("page two, one item per page", func(t *testing.T) {
			orgs, err := svc.ListOrganizations(ctx, organization.OrganizationListOptions{ListOptions: internal.ListOptions{PageNumber: 2, PageSize: 1}})
			require.NoError(t, err)

			assert.Equal(t, 1, len(orgs.Items))
		})
	})

	t.Run("filter list by names", func(t *testing.T) {
		svc := setup(t, nil)
		want1 := svc.createOrganization(t, ctx)
		want2 := svc.createOrganization(t, ctx)
		_ = svc.createOrganization(t, ctx)

		got, err := svc.ListOrganizations(ctx, organization.OrganizationListOptions{Names: []string{want1.Name, want2.Name}})
		require.NoError(t, err)

		assert.Equal(t, 2, len(got.Items))
		assert.Contains(t, got.Items, want1)
		assert.Contains(t, got.Items, want2)
	})

	t.Run("new user should see zero orgs", func(t *testing.T) {
		svc := setup(t, nil)
		_ = svc.createOrganization(t, ctx)
		_ = svc.createOrganization(t, ctx)

		_, ctx := svc.createUserCtx(t, ctx)

		got, err := svc.ListOrganizations(ctx, organization.OrganizationListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 0, len(got.Items))
	})

	t.Run("get", func(t *testing.T) {
		svc := setup(t, nil)
		want := svc.createOrganization(t, ctx)

		got, err := svc.GetOrganization(ctx, want.Name)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("delete", func(t *testing.T) {
		svc := setup(t, nil)
		sub := svc.createSubscriber(t, ctx)
		org := svc.createOrganization(t, ctx)
		assert.Equal(t, pubsub.NewCreatedEvent(org), <-sub)

		err := svc.DeleteOrganization(ctx, org.Name)
		require.NoError(t, err)
		assert.Equal(t, pubsub.NewDeletedEvent(&organization.Organization{ID: org.ID}), <-sub)

		_, err = svc.GetOrganization(ctx, org.Name)
		assert.Equal(t, internal.ErrResourceNotFound, err)
	})

	t.Run("delete non-existent org", func(t *testing.T) {
		svc := setup(t, nil)

		err := svc.DeleteOrganization(ctx, "does-not-exist")
		assert.Equal(t, internal.ErrResourceNotFound, err)
	})
}
