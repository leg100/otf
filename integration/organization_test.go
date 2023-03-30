package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})

	t.Run("create", func(t *testing.T) {
		svc := setup(t, nil)
		org, err := svc.CreateOrganization(ctx, organization.OrganizationCreateOptions{
			Name: otf.String(uuid.NewString()),
		})
		require.NoError(t, err)

		t.Run("duplicate error", func(t *testing.T) {
			_, err := svc.CreateOrganization(ctx, organization.OrganizationCreateOptions{
				Name: otf.String(org.Name),
			})
			require.Equal(t, otf.ErrResourceAlreadyExists, err)
		})
	})

	t.Run("update name", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)

		want := uuid.NewString()
		org, err := svc.UpdateOrganization(ctx, org.Name, organization.OrganizationUpdateOptions{
			Name: otf.String(want),
		})
		require.NoError(t, err)

		assert.Equal(t, want, org.Name)
	})

	t.Run("list with pagination", func(t *testing.T) {
		svc := setup(t, nil)
		_ = svc.createOrganization(t, ctx)
		_ = svc.createOrganization(t, ctx)

		t.Run("page one, two items per page", func(t *testing.T) {
			orgs, err := svc.ListOrganizations(ctx, organization.OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 2}})
			require.NoError(t, err)

			assert.Equal(t, 2, len(orgs.Items))
		})

		t.Run("page one, one item per page", func(t *testing.T) {
			orgs, err := svc.ListOrganizations(ctx, organization.OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 1}})
			require.NoError(t, err)

			assert.Equal(t, 1, len(orgs.Items))
		})

		t.Run("page two, one item per page", func(t *testing.T) {
			orgs, err := svc.ListOrganizations(ctx, organization.OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 2, PageSize: 1}})
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

	t.Run("get", func(t *testing.T) {
		svc := setup(t, nil)
		want := svc.createOrganization(t, ctx)

		got, err := svc.GetOrganization(ctx, want.Name)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("delete", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)

		err := svc.DeleteOrganization(ctx, org.Name)
		require.NoError(t, err)

		_, err = svc.GetOrganization(ctx, org.Name)
		assert.Equal(t, otf.ErrResourceNotFound, err)
	})

	t.Run("delete non-existent org", func(t *testing.T) {
		svc := setup(t, nil)

		err := svc.DeleteOrganization(ctx, "does-not-exist")
		assert.Equal(t, otf.ErrResourceNotFound, err)
	})
}
