package organization

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/require"
)

type (
	fakeService struct {
		orgs []*Organization

		Service
	}
	unprivilegedSubject struct {
		internal.Subject
	}
)

func NewTestOrganization(t *testing.T) *Organization {
	return &Organization{
		Name: uuid.NewString(),
	}
}

func (f *fakeService) ListOrganizations(ctx context.Context, opts ListOptions) (*resource.Page[*Organization], error) {
	return &resource.Page[*Organization]{
		Items:      f.orgs,
		Pagination: resource.NewPagination(opts.PageOptions, int64(len(f.orgs))),
	}, nil
}

func (f *fakeService) DeleteOrganization(context.Context, string) error {
	return nil
}

func (s *unprivilegedSubject) CanAccessSite(_ rbac.Action) bool {
	return false
}

func newFakeWeb(t *testing.T, svc *fakeService, restrict bool) *web {
	renderer, err := html.NewRenderer(false)
	require.NoError(t, err)
	return &web{
		svc:                          svc,
		Renderer:                     renderer,
		RestrictOrganizationCreation: restrict,
	}
}
