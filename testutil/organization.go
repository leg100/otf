package testutil

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/require"
)

func NewOrganizationService(t *testing.T, db otf.DB) *organization.Service {
	service := organization.NewService(organization.Options{
		DB:     db,
		Logger: logr.Discard(),
	})
	return service
}

func CreateOrganization(t *testing.T, db otf.DB) *organization.Organization {
	ctx := context.Background()
	svc := NewOrganizationService(t, db)
	org, err := svc.CreateOrganization(ctx, otf.OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		svc.DeleteOrganization(ctx, org.Name())
	})
	return org
}
