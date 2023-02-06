package organization

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestOrganization(t *testing.T) otf.Organization {
	return newTestOrganization(t)
}

func CreateTestOrganization(t *testing.T, db otf.DB) otf.Organization {
	ctx := context.Background()
	org := newTestOrganization(t)
	orgDB := NewDB(db)
	err := orgDB.create(ctx, org)
	require.NoError(t, err)

	t.Cleanup(func() {
		orgDB.delete(ctx, org.Name())
	})
	return org
}

func newTestOrganization(t *testing.T) *Organization {
	opts := OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	}
	org, err := NewOrganization(opts)
	require.NoError(t, err)

	return org
}
