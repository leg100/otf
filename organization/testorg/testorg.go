// Package testorg provides test helpers for organizations to packages outside
// organization (avoids the import cycle ./sql -> ./organization -> ./sql)
package testorg

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/require"
)

func NewTestOrganization(t *testing.T) *organization.Organization {
	org, err := organization.NewOrganization(organization.OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)
	return org
}

func CreateTestOrganization(t *testing.T, db otf.DB) *organization.Organization {
	ctx := context.Background()
	org := NewTestOrganization(t)
	orgDB := organization.NewDB(db)
	err := orgDB.CreateOrganization(ctx, org)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteOrganization(ctx, org.Name())
	})
	return org
}
