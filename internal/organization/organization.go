// Package organization is responsible for OTF organizations
package organization

import (
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
)

type (
	// Organization is an OTF organization, comprising workspaces, users, etc.
	Organization struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Name      string    `json:"name"`

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		Email                      *string
		CollaboratorAuthPolicy     *string
		SessionRemember            *int
		SessionTimeout             *int
		AllowForceDeleteWorkspaces bool
	}

	// OrganizationList represents a list of Organizations.
	OrganizationList struct {
		*resource.Pagination
		Items []*Organization
	}

	// ListOptions represents the options for listing organizations.
	OrganizationListOptions struct {
		resource.ListOptions
	}

	// UpdateOptions represents the options for updating an
	// organization.
	OrganizationUpdateOptions struct {
		Name            *string
		SessionRemember *int
		SessionTimeout  *int

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		Email                      *string
		CollaboratorAuthPolicy     *string
		AllowForceDeleteWorkspaces *bool
	}
)

func (org *Organization) String() string { return org.ID }

func (org *Organization) Update(opts OrganizationUpdateOptions) error {
	if opts.Name != nil {
		org.Name = *opts.Name
	}
	if opts.Email != nil {
		org.Email = opts.Email
	}
	if opts.CollaboratorAuthPolicy != nil {
		org.CollaboratorAuthPolicy = opts.CollaboratorAuthPolicy
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
	org.UpdatedAt = internal.CurrentTimestamp()
	return nil
}
