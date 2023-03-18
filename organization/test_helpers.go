package organization

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestService(t *testing.T, db otf.DB) *service {
	service := NewService(Options{
		DB:        db,
		Logger:    logr.Discard(),
		Publisher: &otf.FakePublisher{},
	})
	service.Authorizer = otf.NewAllowAllAuthorizer()
	service.site = otf.NewAllowAllAuthorizer()
	return service
}

func NewTestOrganization(t *testing.T) *Organization {
	org, err := NewOrganization(OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)
	return org
}

func CreateTestOrganization(t *testing.T, db otf.DB) *Organization {
	ctx := context.Background()
	svc := NewTestService(t, db)
	org, err := svc.CreateOrganization(ctx, OrganizationCreateOptions{
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
		db.delete(ctx, org.Name)
	})
	return org
}

type fakeService struct {
	orgs []*Organization

	Service
}

func (f *fakeService) CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error) {
	return NewOrganization(opts)
}

func (f *fakeService) ListOrganizations(ctx context.Context, opts OrganizationListOptions) (*OrganizationList, error) {
	return &OrganizationList{
		Items:      f.orgs,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.orgs)),
	}, nil
}
