package orgcreator

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/stretchr/testify/require"
)

func NewTestOrganization(t *testing.T) *organization.Organization {
	org, err := NewOrganization(OrganizationCreateOptions{
		Name: internal.String(uuid.NewString()),
	})
	require.NoError(t, err)
	return org
}

type fakeService struct {
	Service
}

func (f *fakeService) CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (*organization.Organization, error) {
	return NewOrganization(opts)
}
