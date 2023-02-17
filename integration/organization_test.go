package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization_Create(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	svc := testutil.NewOrganizationService(t, db)
	name := uuid.NewString()

	t.Cleanup(func() {
		svc.DeleteOrganization(ctx, name)
	})

	_, err := svc.CreateOrganization(ctx, otf.OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)

	t.Run("duplicate error", func(t *testing.T) {
		_, err := svc.CreateOrganization(ctx, otf.OrganizationCreateOptions{
			Name: otf.String(uuid.NewString()),
		})
		require.Equal(t, otf.ErrResourceAlreadyExists, err)
	})
}

func TestOrganization_Update(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	svc := testutil.NewOrganizationService(t, db)
	org := testutil.CreateOrganization(t, db)

	// update org name
	want := uuid.NewString()
	org, err := svc.UpdateOrganization(ctx, org.Name(), organization.UpdateOptions{
		Name: &want,
	})
	require.NoError(t, err)

	assert.Equal(t, want, org.Name())
}
