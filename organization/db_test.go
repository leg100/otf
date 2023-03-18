package organization

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization_Create(t *testing.T) {
	ctx := context.Background()
	db := &pgdb{sql.NewTestDB(t)}

	t.Run("create", func(t *testing.T) {
		org := NewTestOrganization(t)

		t.Cleanup(func() {
			db.delete(ctx, org.Name)
		})

		err := db.create(ctx, org)
		require.NoError(t, err)

		t.Run("duplicate error", func(t *testing.T) {
			err := db.create(ctx, org)
			require.Equal(t, otf.ErrResourceAlreadyExists, err)
		})
	})

	t.Run("update name", func(t *testing.T) {
		org := createTestOrganization(t, db)

		want := uuid.NewString()
		org, err := db.update(ctx, org.Name, func(org *Organization) error {
			org.Name = want
			return nil
		})
		require.NoError(t, err)

		assert.Equal(t, want, org.Name)
	})

	t.Run("list with pagination", func(t *testing.T) {
		_ = createTestOrganization(t, db)
		_ = createTestOrganization(t, db)

		t.Run("page one, two items per page", func(t *testing.T) {
			orgs, err := db.list(ctx, OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 2}})
			require.NoError(t, err)

			assert.Equal(t, 2, len(orgs.Items))
		})

		t.Run("page one, one item per page", func(t *testing.T) {
			orgs, err := db.list(ctx, OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 1}})
			require.NoError(t, err)

			assert.Equal(t, 1, len(orgs.Items))
		})

		t.Run("page two, one item per page", func(t *testing.T) {
			orgs, err := db.list(ctx, OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 2, PageSize: 1}})
			require.NoError(t, err)

			assert.Equal(t, 1, len(orgs.Items))
		})
	})

	t.Run("filter list by names", func(t *testing.T) {
		want1 := createTestOrganization(t, db)
		want2 := createTestOrganization(t, db)
		_ = createTestOrganization(t, db)

		got, err := db.list(ctx, OrganizationListOptions{Names: []string{want1.Name, want2.Name}})
		require.NoError(t, err)

		assert.Equal(t, 2, len(got.Items))
		assert.Contains(t, got.Items, want1)
		assert.Contains(t, got.Items, want2)

	})

	t.Run("get", func(t *testing.T) {
		want := createTestOrganization(t, db)

		got, err := db.get(ctx, want.Name)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("delete", func(t *testing.T) {
		org := createTestOrganization(t, db)

		err := db.delete(ctx, org.Name)
		require.NoError(t, err)

		_, err = db.get(ctx, org.Name)
		assert.Equal(t, otf.ErrResourceNotFound, err)
	})

	t.Run("delete non-existent org", func(t *testing.T) {
		err := db.delete(ctx, "does-not-exist")
		assert.Equal(t, otf.ErrResourceNotFound, err)
	})
}
