// Package organization is responsible for OTF organizations
package organization

import (
	"time"

	"github.com/leg100/otf/internal"
)

type (
	// Organization is an OTF organization, comprising workspaces, users, etc.
	Organization struct {
		ID              string    `json:"id"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
		Name            string    `json:"name"`
		SessionRemember int       `json:"session_remember"`
		SessionTimeout  int       `json:"session_timeout"`
	}

	// OrganizationList represents a list of Organizations.
	OrganizationList struct {
		*internal.Pagination
		Items []*Organization
	}

	// ListOptions represents the options for listing organizations.
	OrganizationListOptions struct {
		internal.ListOptions
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
