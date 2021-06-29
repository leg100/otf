package ots

import (
	"errors"
	"fmt"

	tfe "github.com/leg100/go-tfe"
)

const (
	DefaultSessionTimeout         = 20160
	DefaultSessionExpiration      = 20160
	DefaultCollaboratorAuthPolicy = "password"
	DefaultCostEstimationEnabled  = true
)

type OrganizationList struct {
	*Pagination
	Items []*tfe.Organization
}

// OrganizationListOptions represents the options for listing organizations.
type OrganizationListOptions struct {
	ListOptions
}

type OrganizationService interface {
	CreateOrganization(opts *tfe.OrganizationCreateOptions) (*tfe.Organization, error)
	GetOrganization(name string) (*tfe.Organization, error)
	ListOrganizations(opts OrganizationListOptions) (*OrganizationList, error)
	UpdateOrganization(name string, opts *tfe.OrganizationUpdateOptions) (*tfe.Organization, error)
	DeleteOrganization(name string) error
	GetEntitlements(name string) (*tfe.Entitlements, error)
}

func NewOrganizationFromOptions(opts *tfe.OrganizationCreateOptions, overrides ...func(*tfe.Organization)) (*tfe.Organization, error) {
	if !validString(opts.Name) {
		return nil, tfe.ErrRequiredName
	}
	if !validStringID(opts.Name) {
		return nil, tfe.ErrInvalidName
	}
	if !validString(opts.Email) {
		return nil, errors.New("email is required")
	}

	org := &tfe.Organization{
		Name:                   *opts.Name,
		Email:                  *opts.Email,
		Permissions:            &tfe.OrganizationPermissions{},
		SessionTimeout:         DefaultSessionTimeout,
		SessionRemember:        DefaultSessionExpiration,
		CollaboratorAuthPolicy: DefaultCollaboratorAuthPolicy,
		CostEstimationEnabled:  DefaultCostEstimationEnabled,
	}

	for _, or := range overrides {
		or(org)
	}

	if opts.SessionTimeout != nil {
		org.SessionTimeout = *opts.SessionTimeout
	}

	if opts.SessionRemember != nil {
		org.SessionRemember = *opts.SessionRemember
	}

	if opts.CollaboratorAuthPolicy != nil {
		org.CollaboratorAuthPolicy = *opts.CollaboratorAuthPolicy
	}

	if opts.CostEstimationEnabled != nil {
		org.CostEstimationEnabled = *opts.CostEstimationEnabled
	}

	return org, nil
}

func UpdateOrganizationFromOptions(org *tfe.Organization, opts *tfe.OrganizationUpdateOptions) error {
	if opts.Name != nil {
		if !validStringID(opts.Name) {
			return tfe.ErrInvalidName
		}
		org.Name = *opts.Name
	}
	if opts.Email != nil {
		org.Email = *opts.Email
	}
	if opts.SessionTimeout != nil {
		org.SessionTimeout = *opts.SessionTimeout
	}
	if opts.SessionRemember != nil {
		org.SessionRemember = *opts.SessionRemember
	}
	if opts.CollaboratorAuthPolicy != nil {
		org.CollaboratorAuthPolicy = *opts.CollaboratorAuthPolicy
	}
	if opts.CostEstimationEnabled != nil {
		org.CostEstimationEnabled = *opts.CostEstimationEnabled
	}
	if opts.OwnersTeamSAMLRoleID != nil {
		org.OwnersTeamSAMLRoleID = *opts.OwnersTeamSAMLRoleID
	}

	return nil
}

// We currently only support State Storage...
func DefaultEntitlements(orgID string) *tfe.Entitlements {
	return &tfe.Entitlements{
		ID:           orgID,
		StateStorage: true,
	}
}

func NewOrganizationID() string {
	return fmt.Sprintf("org-%s", GenerateRandomString(16))
}
