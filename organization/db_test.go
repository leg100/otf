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

func newTestDB(t *testing.T) *pgdb {
	return newDB(sql.NewTestDB(t))
}

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

func TestOrganization_Create(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := newTestOrganization(t)

	t.Cleanup(func() {
		db.delete(ctx, org.Name())
	})

	err := db.create(ctx, org)
	require.NoError(t, err)

	t.Run("Duplicate", func(t *testing.T) {
		err := db.create(ctx, org)
		require.Equal(t, otf.ErrResourceAlreadyExists, err)
	})
}

func TestOrganization_Update(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	newName := uuid.NewString()
	org, err := db.update(ctx, org.Name(), func(org *Organization) error {
		org.Update(updateOptions{Name: &newName})
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, newName, org.Name())
}

func TestOrganization_Get(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	t.Run("by name", func(t *testing.T) {
		got, err := db.get(ctx, org.Name())
		require.NoError(t, err)

		assert.Equal(t, org.Name(), got.Name())
		assert.Equal(t, org.ID(), got.ID())
	})

	t.Run("by id", func(t *testing.T) {
		got, err := db.getByID(ctx, org.ID())
		require.NoError(t, err)

		assert.Equal(t, org.Name(), got.Name())
		assert.Equal(t, org.ID(), got.ID())
	})
}

func TestOrganization_List(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	ol, err := db.list(context.Background(), listOptions{})
	require.NoError(t, err)

	assert.Contains(t, ol.Items, org)
}

func TestOrganization_ListWithPagination(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	_ = createTestOrganization(t, db)
	_ = createTestOrganization(t, db)

	t.Run("page one, two items per page", func(t *testing.T) {
		orgs, err := db.list(ctx, listOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 2}})
		require.NoError(t, err)

		assert.Equal(t, 2, len(orgs.Items))
	})

	t.Run("page one, one item per page", func(t *testing.T) {
		orgs, err := db.list(ctx, listOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 1}})
		require.NoError(t, err)

		assert.Equal(t, 1, len(orgs.Items))
	})

	t.Run("page two, one item per page", func(t *testing.T) {
		orgs, err := db.list(ctx, listOptions{ListOptions: otf.ListOptions{PageNumber: 2, PageSize: 1}})
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
