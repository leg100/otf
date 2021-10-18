package sql

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization_Create(t *testing.T) {
	db := NewOrganizationDB(newTestDB(t))

	org, err := db.Create(newTestOrganization())
	require.NoError(t, err)

	db.Delete(org.Name)
}

func TestOrganization_Update(t *testing.T) {
	db := newTestDB(t)
	odb := NewOrganizationDB(db)
	org := createTestOrganization(t, db)

	org, err := odb.Update(org.Name, func(org *otf.Organization) error {
		org.Email = "newguy@automatize.co.uk"
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, "newguy@automatize.co.uk", org.Email)
}

func TestOrganization_Get(t *testing.T) {
	db := newTestDB(t)
	odb := NewOrganizationDB(db)
	org := createTestOrganization(t, db)

	got, err := odb.Get(org.Name)
	require.NoError(t, err)

	assert.Equal(t, org.Name, got.Name)
}

func TestOrganization_List(t *testing.T) {
	db := newTestDB(t)
	odb := NewOrganizationDB(db)
	org := createTestOrganization(t, db)

	ol, err := odb.List(otf.OrganizationListOptions{})
	require.NoError(t, err)

	assert.Contains(t, ol.Items, org)
}

func TestOrganization_ListWithPagination(t *testing.T) {
	db := newTestDB(t)
	odb := NewOrganizationDB(db)
	_ = createTestOrganization(t, db)
	_ = createTestOrganization(t, db)

	orgs, err := odb.List(otf.OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 2}})
	require.NoError(t, err)

	assert.Equal(t, 2, len(orgs.Items))

	orgs, err = odb.List(otf.OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 1}})
	require.NoError(t, err)

	assert.Equal(t, 1, len(orgs.Items))

	orgs, err = odb.List(otf.OrganizationListOptions{ListOptions: otf.ListOptions{PageNumber: 2, PageSize: 1}})
	require.NoError(t, err)

	assert.Equal(t, 1, len(orgs.Items))
}

func TestOrganization_Delete(t *testing.T) {
	db := newTestDB(t)
	odb := NewOrganizationDB(db)
	org := createTestOrganization(t, db)

	require.NoError(t, odb.Delete(org.Name))

	_, err := odb.Get(org.Name)
	assert.Equal(t, otf.ErrResourceNotFound, err)
}

func TestOrganization_DeleteError(t *testing.T) {
	db := newTestDB(t)
	odb := NewOrganizationDB(db)
	_ = createTestOrganization(t, db)

	err := odb.Delete("non-existent-org")

	assert.Equal(t, otf.ErrResourceNotFound, err)
}
