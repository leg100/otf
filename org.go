package ots

import (
	"fmt"

	tfe "github.com/leg100/go-tfe"
)

const (
	DefaultSessionTimeout         = 20160
	DefaultSessionExpiration      = 20160
	DefaultCollaboratorAuthPolicy = "password"
	DefaultCostEstimationEnabled  = true
)

var (
	DefaultOrganizationPermissions = tfe.OrganizationPermissions{
		CanCreateWorkspace: true,
		CanUpdate:          true,
		CanDestroy:         true,
	}
)

type OrganizationService interface {
	CreateOrganization(opts *tfe.OrganizationCreateOptions) (*tfe.Organization, error)
	GetOrganization(name string) (*tfe.Organization, error)
	ListOrganizations(opts tfe.OrganizationListOptions) (*tfe.OrganizationList, error)
	UpdateOrganization(name string, opts *tfe.OrganizationUpdateOptions) (*tfe.Organization, error)
	DeleteOrganization(name string) error
	GetEntitlements(name string) (*tfe.Entitlements, error)
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
