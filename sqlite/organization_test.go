package sqlite

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization_Create(t *testing.T) {
	db := NewOrganizationDB(newTestDB(t))

	run, err := db.Create(newTestOrganization("org-123", "automatize"))
	require.NoError(t, err)

	assert.Equal(t, int64(1), run.Model.ID)
}

func TestOrganization_Update(t *testing.T) {
	db := newTestDB(t)
	odb := NewOrganizationDB(db)
	_ = createTestOrganization(t, db, "org-123", "automatize")

	org, err := odb.Update("automatize", func(org *otf.Organization) error {
		org.Email = "newguy@automatize.co.uk"
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, "newguy@automatize.co.uk", org.Email)
}

func TestOrganization_Get(t *testing.T) {
	db := newTestDB(t)
	odb := NewOrganizationDB(db)
	_ = createTestOrganization(t, db, "org-123", "automatize")

	org, err := odb.Get("automatize")
	require.NoError(t, err)

	assert.Equal(t, "automatize", org.Name)
}

func TestOrganization_List(t *testing.T) {
	db := newTestDB(t)
	odb := NewOrganizationDB(db)
	_ = createTestOrganization(t, db, "org-123", "automatize")

	orgs, err := odb.List(otf.OrganizationListOptions{})
	require.NoError(t, err)

	require.Equal(t, 1, len(orgs.Items))
}

func TestOrganization_ListWithPagination(t *testing.T) {
	db := newTestDB(t)
	odb := NewOrganizationDB(db)
	_ = createTestOrganization(t, db, "org-123", "automatize")
	_ = createTestOrganization(t, db, "org-456", "pinasonic")

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
	_ = createTestOrganization(t, db, "org-123", "automatize")

	require.NoError(t, odb.Delete("automatize"))

	orgs, err := odb.List(otf.OrganizationListOptions{})
	require.NoError(t, err)

	assert.Equal(t, 0, len(orgs.Items))
}

func TestOrganization_DeleteError(t *testing.T) {
	db := newTestDB(t)
	odb := NewOrganizationDB(db)
	_ = createTestOrganization(t, db, "org-123", "automatize")

	err := odb.Delete("non-existent-org")

	assert.Equal(t, otf.ErrResourceNotFound, err)
}
