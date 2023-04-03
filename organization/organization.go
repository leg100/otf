// Package organization is responsible for OTF organizations
package organization

import (
	"time"

	"github.com/leg100/otf"
)

type (
	// Organization is an OTF organization, comprising workspaces, users, etc.
	Organization struct {
		ID              string
		CreatedAt       time.Time
		UpdatedAt       time.Time
		Name            string
		SessionRemember int
		SessionTimeout  int
	}

	// OrganizationList represents a list of Organizations.
	OrganizationList struct {
		*otf.Pagination
		Items []*Organization
	}

	// ListOptions represents the options for listing organizations.
	OrganizationListOptions struct {
		Names []string // filter organizations by name
		otf.ListOptions
	}

	// UpdateOptions represents the options for updating an
	// organization.
	OrganizationUpdateOptions struct {
		Name            *string
		SessionRemember *int
		SessionTimeout  *int
	}
)

func (org *Organization) String() string { return org.ID }

func (org *Organization) Update(opts OrganizationUpdateOptions) error {
	if opts.Name != nil {
		org.Name = *opts.Name
	}
	if opts.SessionTimeout != nil {
		org.SessionTimeout = *opts.SessionTimeout
	}
	if opts.SessionRemember != nil {
		org.SessionRemember = *opts.SessionRemember
	}
	return nil
}
