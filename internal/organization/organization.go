// Package organization is responsible for OTF organizations
package organization

import (
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
)

const (
	DefaultSessionTimeout    = 20160
	DefaultSessionExpiration = 20160
)

type (
	// Organization is an OTF organization, comprising workspaces, users, etc.
	Organization struct {
		resource.ID `jsonapi:"primary,organizations"`

		CreatedAt time.Time `jsonapi:"attribute" json:"created-at"`
		UpdatedAt time.Time `jsonapi:"attribute" json:"updated-at"`
		Name      string    `jsonapi:"attribute" json:"name"`

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		Email                      *string
		CollaboratorAuthPolicy     *string
		SessionRemember            *int
		SessionTimeout             *int
		AllowForceDeleteWorkspaces bool
		CostEstimationEnabled      bool
	}

	// UpdateOptions represents the options for updating an organization.
	UpdateOptions struct {
		Name            *string
		SessionRemember *int
		SessionTimeout  *int

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		Email                      *string
		CollaboratorAuthPolicy     *string
		CostEstimationEnabled      *bool
		AllowForceDeleteWorkspaces *bool
	}

	// CreateOptions represents the options for creating an organization. See
	// types.CreateOptions for more details.
	CreateOptions struct {
		Name *string

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		Email                      *string
		CollaboratorAuthPolicy     *string
		CostEstimationEnabled      *bool
		SessionRemember            *int
		SessionTimeout             *int
		AllowForceDeleteWorkspaces *bool
	}
)

func NewOrganization(opts CreateOptions) (*Organization, error) {
	if err := resource.ValidateName(opts.Name); err != nil {
		return nil, err
	}
	org := Organization{
		Name:                   *opts.Name,
		CreatedAt:              internal.CurrentTimestamp(nil),
		UpdatedAt:              internal.CurrentTimestamp(nil),
		ID:                     resource.NewID("org"),
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
	if opts.CostEstimationEnabled != nil {
		org.CostEstimationEnabled = *opts.CostEstimationEnabled
	}
	return &org, nil
}

func (org *Organization) Update(opts UpdateOptions) error {
	if opts.Name != nil {
		org.Name = *opts.Name
	}
	if opts.Email != nil {
		org.Email = opts.Email
	}
	if opts.CollaboratorAuthPolicy != nil {
		org.CollaboratorAuthPolicy = opts.CollaboratorAuthPolicy
	}
	if opts.CostEstimationEnabled != nil {
		org.CostEstimationEnabled = *opts.CostEstimationEnabled
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
	org.UpdatedAt = internal.CurrentTimestamp(nil)
	return nil
}
