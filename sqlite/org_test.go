package sqlite

import (
	"testing"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOrganization(t *testing.T) {
	db, err := New(":memory:", &gorm.Config{})
	require.NoError(t, err)

	svc := NewOrganizationDB(db)

	// Create

	org, err := svc.Create(&ots.Organization{
		Name:       "automatize",
		ExternalID: "org-123",
		Email:      "sysadmin@automatize.co.uk",
	})
	require.NoError(t, err)

	require.Equal(t, uint(1), org.InternalID)
	require.Equal(t, "automatize", org.Name)
	require.Equal(t, "sysadmin@automatize.co.uk", org.Email)

	// Create second org

	org, err = svc.Create(&ots.Organization{
		Name:       "second",
		ExternalID: "org-456",
		Email:      "sysadmin@second.org",
	})
	require.NoError(t, err)

	require.Equal(t, uint(2), org.InternalID)
	require.Equal(t, "second", org.Name)
	require.Equal(t, "sysadmin@second.org", org.Email)

	// Update

	org, err = svc.Update("automatize", func(org *ots.Organization) error {
		org.Email = "newguy@automatize.co.uk"
		return nil
	})
	require.NoError(t, err)

	require.Equal(t, "newguy@automatize.co.uk", org.Email)

	// Get

	org, err = svc.Get("automatize")
	require.NoError(t, err)

	require.Equal(t, "newguy@automatize.co.uk", org.Email)

	// List

	orgs, err := svc.List(tfe.OrganizationListOptions{})
	require.NoError(t, err)

	require.Equal(t, 2, len(orgs.Items))

	// List with pagination

	orgs, err = svc.List(tfe.OrganizationListOptions{ListOptions: tfe.ListOptions{PageNumber: 1, PageSize: 1}})
	require.NoError(t, err)

	require.Equal(t, 1, len(orgs.Items))

	orgs, err = svc.List(tfe.OrganizationListOptions{ListOptions: tfe.ListOptions{PageNumber: 2, PageSize: 1}})
	require.NoError(t, err)

	require.Equal(t, 1, len(orgs.Items))

	// Delete

	require.NoError(t, svc.Delete("automatize"))
}
