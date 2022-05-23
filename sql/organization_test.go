package sql

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization_Create(t *testing.T) {
	db := newTestDB(t)
	org := newTestOrganization()

	t.Cleanup(func() {
		db.OrganizationStore().Delete(org.Name())
	})

	_, err := db.OrganizationStore().Create(org)
	require.NoError(t, err)

	t.Run("Duplicate", func(t *testing.T) {
		_, err := db.OrganizationStore().Create(org)
		require.Equal(t, otf.ErrResourcesAlreadyExists, err)
	})
}

func TestOrganization_Update(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	newName := uuid.NewString()
	org, err := db.OrganizationStore().Update(org.Name(), func(org *otf.Organization) error {
		otf.UpdateOrganizationFromOpts(org, otf.OrganizationUpdateOptions{Name: &newName})
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, newName, org.Name())
}

func TestOrganization_Get(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	got, err := db.OrganizationStore().Get(org.Name())
	require.NoError(t, err)

	assert.Equal(t, org.Name(), got.Name())
	assert.Equal(t, org.ID(), got.ID())
}

func TestOrganization_List(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	ol, err := db.OrganizationStore().List(otf.OrganizationListOptions{})
	require.NoError(t, err)

	assert.Contains(t, ol.Items, org)
}

func TestOrganization_ListWithPagination(t *testing.T) {
	db := newTestDB(t)
	_ = createTestOrganization(t, db)
	_ = createTestOrganization(t, db)

	t.Run("page one, two items per page", func(t *testing.T) {
		orgs, err := db.OrganizationStore().List(otf.OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 2}})
		require.NoError(t, err)

		assert.Equal(t, 2, len(orgs.Items))
	})

	t.Run("page one, one item per page", func(t *testing.T) {
		orgs, err := db.OrganizationStore().List(otf.OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 1}})
		require.NoError(t, err)

		assert.Equal(t, 1, len(orgs.Items))
	})

	t.Run("page two, one item per page", func(t *testing.T) {
		orgs, err := db.OrganizationStore().List(otf.OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 2, PageSize: 1}})
		require.NoError(t, err)

		assert.Equal(t, 1, len(orgs.Items))
	})
}

func TestOrganization_Delete(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	require.NoError(t, db.OrganizationStore().Delete(org.Name()))

	_, err := db.OrganizationStore().Get(org.Name())
	assert.Equal(t, otf.ErrResourceNotFound, err)
}

func TestOrganization_DeleteError(t *testing.T) {
	db := newTestDB(t)
	_ = createTestOrganization(t, db)

	err := db.OrganizationStore().Delete("non-existent-org")

	assert.Equal(t, otf.ErrResourceNotFound, err)
}
