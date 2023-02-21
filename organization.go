package otf

import (
	"context"
	"errors"
	"fmt"

	"github.com/leg100/otf/http/jsonapi"
)

type Organization struct {
	Name string
}

type OrganizationList struct {
	*Pagination
	Items []Organization
}

type OrganizationService interface {
	CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (Organization, error)
	EnsureCreatedOrganization(ctx context.Context, opts OrganizationCreateOptions) (Organization, error)
	GetOrganization(ctx context.Context, name string) (Organization, error)
	GetOrganizationJSONAPI(ctx context.Context, name string) (*jsonapi.Organization, error)
}

// OrganizationCreateOptions represents the options for creating an
// organization. See dto.OrganizationCreateOptions for more details.
type OrganizationCreateOptions struct {
	Name            *string `schema:"name,required"`
	SessionRemember *int
	SessionTimeout  *int
}

func (opts *OrganizationCreateOptions) Validate() error {
	if opts.Name == nil {
		return errors.New("name required")
	}
	if *opts.Name == "" {
		return errors.New("name cannot be empty")
	}
	if !ValidStringID(opts.Name) {
		return fmt.Errorf("invalid name: %s", *opts.Name)
	}
	return nil
}
