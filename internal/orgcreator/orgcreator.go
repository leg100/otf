// Package orgcreator is responsible for the creation of organizations. This
// would have been the responsibility of package 'organization' but carving out
// this particular responsibility into a separate package avoids an import cycle
// with the 'auth' package.
package orgcreator

import (
	"errors"
	"fmt"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
)

const (
	DefaultSessionTimeout    = 20160
	DefaultSessionExpiration = 20160
)

type (
	// OrganizationCreateOptions represents the options for creating an
	// organization. See dto.OrganizationCreateOptions for more details.
	OrganizationCreateOptions struct {
		Name *string `schema:"name,required"`

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		Email                      *string
		CollaboratorAuthPolicy     *string
		SessionRemember            *int
		SessionTimeout             *int
		AllowForceDeleteWorkspaces *bool
	}
)

func NewOrganization(opts OrganizationCreateOptions) (*organization.Organization, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	org := organization.Organization{
		Name:                   *opts.Name,
		CreatedAt:              internal.CurrentTimestamp(),
		UpdatedAt:              internal.CurrentTimestamp(),
		ID:                     internal.NewID("org"),
		Email:                  opts.Email,
		CollaboratorAuthPolicy: opts.CollaboratorAuthPolicy,
	}
	if opts.SessionTimeout != nil {
		org.SessionTimeout = opts.SessionTimeout
	}
	if opts.SessionRemember != nil {
		org.SessionRemember = opts.SessionRemember
	}
	if opts.AllowForceDeleteWorkspaces != nil {
		org.AllowForceDeleteWorkspaces = *opts.AllowForceDeleteWorkspaces
	}
	return &org, nil
}

func (opts *OrganizationCreateOptions) Validate() error {
	if opts.Name == nil {
		return errors.New("name required")
	}
	if *opts.Name == "" {
		return errors.New("name cannot be empty")
	}
	if !internal.ValidStringID(opts.Name) {
		return fmt.Errorf("invalid name: %s", *opts.Name)
	}
	return nil
}
