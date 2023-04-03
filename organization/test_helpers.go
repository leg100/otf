package organization

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
)

func NewTestOrganization(t *testing.T) *Organization {
	return &Organization{
		Name: uuid.NewString(),
	}
}

type fakeService struct {
	orgs []*Organization

	Service
}

func (f *fakeService) ListOrganizations(ctx context.Context, opts OrganizationListOptions) (*OrganizationList, error) {
	return &OrganizationList{
		Items:      f.orgs,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.orgs)),
	}, nil
}
