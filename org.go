package ots

import (
	"time"

	tfe "github.com/leg100/go-tfe"
	"gorm.io/gorm"
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

// Organization represents a Terraform Enterprise organization.
type Organization struct {
	ID string

	gorm.Model

	Name                   string
	CollaboratorAuthPolicy tfe.AuthPolicyType
	CostEstimationEnabled  bool
	Email                  string
	OwnersTeamSAMLRoleID   string
	Permissions            *tfe.OrganizationPermissions
	SAMLEnabled            bool
	SessionRemember        int
	SessionTimeout         int
	TrialExpiresAt         time.Time
	TwoFactorConformant    bool
}

// OrganizationList represents a list of Organizations.
type OrganizationList struct {
	*tfe.Pagination
	Items []*Organization
}

type OrganizationService interface {
	Create(opts *tfe.OrganizationCreateOptions) (*Organization, error)
	Get(name string) (*Organization, error)
	List(opts tfe.OrganizationListOptions) (*OrganizationList, error)
	Update(name string, opts *tfe.OrganizationUpdateOptions) (*Organization, error)
	Delete(name string) error
	GetEntitlements(name string) (*Entitlements, error)
}

type OrganizationRepository interface {
	Create(org *Organization) (*Organization, error)
	Get(name string) (*Organization, error)
	List(opts tfe.OrganizationListOptions) (*OrganizationList, error)
	Update(name string, fn func(*Organization) error) (*Organization, error)
	Delete(name string) error
}

func NewOrganization(opts *tfe.OrganizationCreateOptions) (*Organization, error) {
	org := Organization{
		Name:                   *opts.Name,
		Email:                  *opts.Email,
		ID:                     GenerateID("org"),
		SessionTimeout:         DefaultSessionTimeout,
		SessionRemember:        DefaultSessionExpiration,
		CollaboratorAuthPolicy: DefaultCollaboratorAuthPolicy,
		CostEstimationEnabled:  DefaultCostEstimationEnabled,
		Permissions:            &DefaultOrganizationPermissions,
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

	return &org, nil
}

func UpdateOrganization(org *Organization, opts *tfe.OrganizationUpdateOptions) (*Organization, error) {
	if opts.Name != nil {
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

	return org, nil
}
