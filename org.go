package ots

import (
	"errors"

	tfe "github.com/hashicorp/go-tfe"
)

const (
	DefaultSessionTimeout         = 20160
	DefaultSessionExpiration      = 20160
	DefaultCollaboratorAuthPolicy = "password"
	DefaultCostEstimationEnabled  = true
)

type OrganizationService interface {
	CreateOrganization(name string, org *tfe.Organization) (*tfe.Organization, error)
	GetOrganization(name string) (*tfe.Organization, error)
	ListOrganizations() ([]*tfe.Organization, error)
	UpdateOrganization(name string, opts *tfe.OrganizationUpdateOptions) (*tfe.Organization, error)
	DeleteOrganization(name string) error
	GetEntitlements(name string) (*tfe.Entitlements, error)
}

func NewOrganizationFromOptions(opts *tfe.OrganizationCreateOptions) (*tfe.Organization, error) {
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
