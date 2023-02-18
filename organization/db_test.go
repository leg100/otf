package organization

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestOrganization(t *testing.T, db *pgdb) *Organization {
	ctx := context.Background()
	org := newTestOrganization(t)
	err := db.create(ctx, org)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.delete(ctx, org.name)
	})
	return org
}

func TestOrganization_List(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	ol, err := db.list(context.Background(), ListOptions{})
	require.NoError(t, err)

	assert.Contains(t, ol.Items, org)
}

func TestOrganization_ListWithPagination(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	_ = createTestOrganization(t, db)
	_ = createTestOrganization(t, db)

	t.Run("page one, two items per page", func(t *testing.T) {
		orgs, err := db.list(ctx, ListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 2}})
		require.NoError(t, err)

		assert.Equal(t, 2, len(orgs.Items))
	})

	t.Run("page one, one item per page", func(t *testing.T) {
		orgs, err := db.list(ctx, ListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 1}})
		require.NoError(t, err)

		assert.Equal(t, 1, len(orgs.Items))
	})

	t.Run("page two, one item per page", func(t *testing.T) {
		orgs, err := db.list(ctx, ListOptions{ListOptions: otf.ListOptions{PageNumber: 2, PageSize: 1}})
		require.NoError(t, err)

		assert.Equal(t, 1, len(orgs.Items))
	})
}

func TestListUserOrganizations(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org1 := createTestOrganization(t, db)
	org2 := createTestOrganization(t, db)
	user := auth.CreateTestUser(t, db,
		otf.WithOrganizationMemberships(org1.Name(), org2.Name()))

	got, err := db.ListOrganizationsByUser(ctx, user.ID())
	require.NoError(t, err)

	assert.Contains(t, got, org1)
	assert.Contains(t, got, org2)
}

func TestOrganization_Delete(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	require.NoError(t, db.delete(ctx, org.name))

	_, err := db.get(ctx, org.name)
	assert.Equal(t, otf.ErrResourceNotFound, err)
}

func TestOrganization_DeleteError(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	_ = createTestOrganization(t, db)

	err := db.delete(ctx, "non-existent-org")

	assert.Equal(t, otf.ErrResourceNotFound, err)
}
