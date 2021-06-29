package sqlite

import (
	"testing"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestOrganization(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"))
	require.NoError(t, err)

	svc := NewOrganizationService(db)

	// Create

	org, err := svc.CreateOrganization(&tfe.OrganizationCreateOptions{
		Name:  ots.String("automatize"),
		Email: ots.String("sysadmin@automatize.co.uk"),
	})
	require.NoError(t, err)

	require.Equal(t, "automatize", org.Name)
	require.Equal(t, "sysadmin@automatize.co.uk", org.Email)

	// Create second org

	org, err = svc.CreateOrganization(&tfe.OrganizationCreateOptions{
		Name:  ots.String("second"),
		Email: ots.String("sysadmin@second.org"),
	})
	require.NoError(t, err)

	require.Equal(t, "second", org.Name)
	require.Equal(t, "sysadmin@second.org", org.Email)

	// Update

	org, err = svc.UpdateOrganization("automatize", &tfe.OrganizationUpdateOptions{
		Email: ots.String("newguy@automatize.co.uk"),
	})
	require.NoError(t, err)

	require.Equal(t, "newguy@automatize.co.uk", org.Email)

	// Get

	org, err = svc.GetOrganization("automatize")
	require.NoError(t, err)

	require.Equal(t, "newguy@automatize.co.uk", org.Email)

	// List

	orgs, err := svc.ListOrganizations(ots.OrganizationListOptions{})
	require.NoError(t, err)

	require.Equal(t, 2, len(orgs.Items))

	// List with pagination

	orgs, err = svc.ListOrganizations(ots.OrganizationListOptions{ListOptions: ots.ListOptions{PageNumber: 1, PageSize: 1}})
	require.NoError(t, err)

	require.Equal(t, 1, len(orgs.Items))

	orgs, err = svc.ListOrganizations(ots.OrganizationListOptions{ListOptions: ots.ListOptions{PageNumber: 2, PageSize: 1}})
	require.NoError(t, err)

	require.Equal(t, 1, len(orgs.Items))

	// Delete

	require.NoError(t, svc.DeleteOrganization("automatize"))
}
