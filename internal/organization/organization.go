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
		ID        resource.TfeID            `jsonapi:"primary,organizations" db:"organization_id"`
		CreatedAt time.Time                 `jsonapi:"attribute" json:"created-at" db:"created_at"`
		UpdatedAt time.Time                 `jsonapi:"attribute" json:"updated-at" db:"updated_at"`
		Name      resource.OrganizationName `jsonapi:"attribute" json:"name" db:"name"`

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		Email                      *string `db:"email"`
		CollaboratorAuthPolicy     *string `db:"collaborator_auth_policy"`
		SessionRemember            *int    `db:"session_remember"`
		SessionTimeout             *int    `db:"session_timeout"`
		AllowForceDeleteWorkspaces bool    `db:"allow_force_delete_workspaces"`
		CostEstimationEnabled      bool    `db:"cost_estimation_enabled"`
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
	if opts.Name == nil {
		return nil, internal.ErrRequiredName
	}
	name, err := resource.NewOrganizationName(*opts.Name)
	if err != nil {
		return nil, err
	}
	org := Organization{
		Name:                   name,
		CreatedAt:              internal.CurrentTimestamp(nil),
		UpdatedAt:              internal.CurrentTimestamp(nil),
		ID:                     resource.NewTfeID(resource.OrganizationKind),
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
		name, err := resource.NewOrganizationName(*opts.Name)
		if err != nil {
			return err
		}
		org.Name = name
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
