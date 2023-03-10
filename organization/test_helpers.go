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
		DB:        db,
		Logger:    logr.Discard(),
		Publisher: &otf.FakePublisher{},
	})
	service.Authorizer = otf.NewAllowAllAuthorizer()
	service.site = otf.NewAllowAllAuthorizer()
	return service
}

func NewTestOrganization(t *testing.T) *otf.Organization {
	org, err := otf.NewOrganization(otf.OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)
	return org
}

func CreateTestOrganization(t *testing.T, db otf.DB) *otf.Organization {
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

func createTestOrganization(t *testing.T, db *pgdb) *otf.Organization {
	ctx := context.Background()
	org := NewTestOrganization(t)
	err := db.create(ctx, org)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.delete(ctx, org.Name)
	})
	return org
}

type fakeService struct {
	orgs []*otf.Organization

	service
}

func (f *fakeService) create(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	return otf.NewOrganization(opts)
}

func (f *fakeService) list(ctx context.Context, opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	return &otf.OrganizationList{
		Items:      f.orgs,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.orgs)),
	}, nil
}
