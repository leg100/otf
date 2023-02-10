package otf

import (
	"context"

	"github.com/leg100/otf/http/jsonapi"
)

type Organization interface {
	Name() string
}

type OrganizationService interface {
	CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (Organization, error)
	EnsureCreatedOrganization(ctx context.Context, opts OrganizationCreateOptions) (Organization, error)
	GetOrganization(ctx context.Context, name string) (Organization, error)
	GetOrganizationJSONAPI(ctx context.Context, name string) (*jsonapi.Organization, error)
}

type OrganizationDB interface {
	GetOrganizationByID(context.Context, string) (Organization, error)
}

// OrganizationCreateOptions represents the options for creating an
// organization. See dto.OrganizationCreateOptions for more details.
type OrganizationCreateOptions struct {
	Name            *string `schema:"name,required"`
	SessionRemember *int
	SessionTimeout  *int
}
