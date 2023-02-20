package organization

import (
	"context"

	"github.com/leg100/otf"
)

type fakeApp struct {
	orgs []*Organization

	app
}

func (f *fakeApp) create(ctx context.Context, opts otf.OrganizationCreateOptions) (*Organization, error) {
	return NewOrganization(opts)
}

func (f *fakeApp) list(ctx context.Context, opts ListOptions) (*OrganizationList, error) {
	return &OrganizationList{
		Items:      f.orgs,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.orgs)),
	}, nil
}
