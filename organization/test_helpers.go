package organization

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestService(t *testing.T, db otf.DB) *Service {
	service := NewService(Options{
		Authorizer: otf.NewAllowAllAuthorizer(),
		DB:         db,
		Logger:     logr.Discard(),
	})
	return service
}

func NewTestOrganization(t *testing.T) *Organization {
	org, err := NewOrganization(otf.OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)
	return org
}

func CreateTestOrganization(t *testing.T, db otf.DB) otf.Organization {
	ctx := context.Background()
	svc := NewTestService(t, db)
	org, err := svc.CreateOrganization(ctx, otf.OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		svc.DeleteOrganization(ctx, org.Name)
	})
	return org
}

func createTestOrganization(t *testing.T, db *pgdb) *Organization {
	ctx := context.Background()
	org := NewTestOrganization(t)
	err := db.create(ctx, org)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.delete(ctx, org.name)
	})
	return org
}
