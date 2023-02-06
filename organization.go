package otf

import (
	"context"
)

type OrganizationService interface {
	CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (Organization, error)
	EnsureCreatedOrganization(ctx context.Context, opts OrganizationCreateOptions) (Organization, error)
	GetOrganization(ctx context.Context, name string) (Organization, error)
}

type Organization interface {
	Name() string
}

// OrganizationCreateOptions represents the options for creating an
// organization. See dto.OrganizationCreateOptions for more details.
type OrganizationCreateOptions struct {
	Name            *string `schema:"name,required"`
	SessionRemember *int
	SessionTimeout  *int
}
