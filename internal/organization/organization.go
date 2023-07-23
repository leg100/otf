// Package organization is responsible for OTF organizations
package organization

import (
	"errors"
	"fmt"
	"time"

	"github.com/leg100/otf/internal"
)

const (
	DefaultSessionTimeout    = 20160
	DefaultSessionExpiration = 20160
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

func NewOrganization(opts OrganizationCreateOptions) (*Organization, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	org := Organization{
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
