package organization

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestOrganization(t *testing.T) otf.Organization {
	opts := OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	}
	org, err := NewOrganization(opts)
	require.NoError(t, err)

	return org
}

func CreateTestOrganization(t *testing.T, db otf.DB) otf.Organization {
	ctx := context.Background()
	org := NewTestOrganization(t)
	orgDB := NewDB(db)
	err := orgDB.CreateOrganization(ctx, org)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteOrganization(ctx, org.Name())
	})
	return org
}
