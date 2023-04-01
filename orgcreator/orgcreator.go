// Package orgcreator is responsible for the creation of organizations. This
// would have been the responsibility of package 'organization' but carving out
// this particular responsibility into a separate package avoids an import cycle
// with the 'auth' package.
package orgcreator

import (
	"errors"
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/organization"
)

const (
	DefaultSessionTimeout    = 20160
	DefaultSessionExpiration = 20160
)

type (
	// OrganizationCreateOptions represents the options for creating an
	// organization. See dto.OrganizationCreateOptions for more details.
	OrganizationCreateOptions struct {
		Name            *string `schema:"name,required"`
		SessionRemember *int
		SessionTimeout  *int
	}
)

func NewOrganization(opts OrganizationCreateOptions) (*organization.Organization, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	org := organization.Organization{
		Name:            *opts.Name,
		CreatedAt:       otf.CurrentTimestamp(),
		UpdatedAt:       otf.CurrentTimestamp(),
		ID:              otf.NewID("org"),
		SessionTimeout:  DefaultSessionTimeout,
		SessionRemember: DefaultSessionExpiration,
	}
	if opts.SessionTimeout != nil {
		org.SessionTimeout = *opts.SessionTimeout
	}
	if opts.SessionRemember != nil {
		org.SessionRemember = *opts.SessionRemember
	}
	return &org, nil
}

func newFromJSONAPI(from jsonapi.Organization) *organization.Organization {
	return &organization.Organization{
		ID:              from.ExternalID,
		CreatedAt:       from.CreatedAt,
		Name:            from.Name,
		SessionRemember: from.SessionRemember,
		SessionTimeout:  from.SessionTimeout,
	}
}

func (opts *OrganizationCreateOptions) Validate() error {
	if opts.Name == nil {
		return errors.New("name required")
	}
	if *opts.Name == "" {
		return errors.New("name cannot be empty")
	}
	if !otf.ValidStringID(opts.Name) {
		return fmt.Errorf("invalid name: %s", *opts.Name)
	}
	return nil
}
