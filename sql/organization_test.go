package sql

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization_Create(t *testing.T) {
	db := newTestDB(t)
	org := newTestOrganization(t)

	t.Cleanup(func() {
		db.DeleteOrganization(context.Background(), org.Name())
	})

	err := db.CreateOrganization(context.Background(), org)
	require.NoError(t, err)

	t.Run("Duplicate", func(t *testing.T) {
		err := db.CreateOrganization(context.Background(), org)
		require.Equal(t, otf.ErrResourcesAlreadyExists, err)
	})
}

func TestOrganization_Update(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	newName := uuid.NewString()
	org, err := db.UpdateOrganization(context.Background(), org.Name(), func(org *otf.Organization) error {
		otf.UpdateOrganizationFromOpts(org, otf.OrganizationUpdateOptions{Name: &newName})
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, newName, org.Name())
}

func TestOrganization_Get(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	got, err := db.GetOrganization(context.Background(), org.Name())
	require.NoError(t, err)

	assert.Equal(t, org.Name(), got.Name())
	assert.Equal(t, org.ID(), got.ID())
}

func TestOrganization_List(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	ol, err := db.ListOrganizations(context.Background(), otf.OrganizationListOptions{})
	require.NoError(t, err)

	assert.Contains(t, ol.Items, org)
}

func TestOrganization_ListWithPagination(t *testing.T) {
	db := newTestDB(t)
	_ = createTestOrganization(t, db)
	_ = createTestOrganization(t, db)

	t.Run("page one, two items per page", func(t *testing.T) {
		orgs, err := db.ListOrganizations(context.Background(), otf.OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 2}})
		require.NoError(t, err)

		assert.Equal(t, 2, len(orgs.Items))
	})

	t.Run("page one, one item per page", func(t *testing.T) {
		orgs, err := db.ListOrganizations(context.Background(), otf.OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 1}})
		require.NoError(t, err)

		assert.Equal(t, 1, len(orgs.Items))
	})

	t.Run("page two, one item per page", func(t *testing.T) {
		orgs, err := db.ListOrganizations(context.Background(), otf.OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 2, PageSize: 1}})
		require.NoError(t, err)

		assert.Equal(t, 1, len(orgs.Items))
	})
}

func TestOrganization_Delete(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	require.NoError(t, db.DeleteOrganization(context.Background(), org.Name()))

	_, err := db.GetOrganization(context.Background(), org.Name())
	assert.Equal(t, otf.ErrResourceNotFound, err)
}

func TestOrganization_DeleteError(t *testing.T) {
	db := newTestDB(t)
	_ = createTestOrganization(t, db)

	err := db.DeleteOrganization(context.Background(), "non-existent-org")

	assert.Equal(t, otf.ErrResourceNotFound, err)
}
