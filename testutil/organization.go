package testutil

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization_Create(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	svc := newOrganizationService(t, db)
	name := uuid.NewString()

	t.Cleanup(func() {
		svc.DeleteOrganization(ctx, name)
	})

	org, err := svc.CreateOrganization(ctx, otf.OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)

	t.Run("duplicate error", func(t *testing.T) {
		err := svc.CreateOrganization(ctx, org)
		require.Equal(t, otf.ErrResourceAlreadyExists, err)
	})
}

func TestOrganization_Update(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	svc := newOrganizationService(t, db)
	org := createOrganization(t, svc)

	newName := uuid.NewString()
	org, err := svc.UpdateOrganization(ctx, org.Name(), func(org *organization.Organization) error {
		org.Update(updateOptions{Name: &newName})
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, newName, org.Name())
}

func newOrganizationService(t *testing.T, db otf.DB) *organization.Service {
	service := organization.NewService(context.Background(), organization.Options{
		Authorizer: &AllowAllAuthorizer{User: &otf.Superuser{"bob"}},
		DB:         db,
		Logger:     logr.Discard(),
	})
	return service
}

func createOrganization(t *testing.T, svc *organization.Service) otf.Organization {
	ctx := context.Background()
	org, err := svc.CreateOrganization(ctx, otf.OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		svc.DeleteOrganization(ctx, org.Name())
	})
	return org
}
