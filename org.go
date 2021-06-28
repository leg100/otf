package ots

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/jsonapi"
	tfe "github.com/hashicorp/go-tfe"
)

const (
	DefaultSessionTimeout         = 20160
	DefaultSessionExpiration      = 20160
	DefaultCollaboratorAuthPolicy = "password"
	DefaultCostEstimationEnabled  = true
)

// Organization represents a Terraform Enterprise organization.
type Organization struct {
	Name                   string                   `jsonapi:"primary,organizations"`
	CollaboratorAuthPolicy tfe.AuthPolicyType       `jsonapi:"attr,collaborator-auth-policy"`
	CostEstimationEnabled  bool                     `jsonapi:"attr,cost-estimation-enabled"`
	CreatedAt              time.Time                `jsonapi:"attr,created-at,iso8601"`
	Email                  string                   `jsonapi:"attr,email"`
	EnterprisePlan         tfe.EnterprisePlanType   `jsonapi:"attr,enterprise-plan"`
	ExternalID             string                   `jsonapi:"attr,external-id"`
	OwnersTeamSAMLRoleID   string                   `jsonapi:"attr,owners-team-saml-role-id"`
	Permissions            *OrganizationPermissions `jsonapi:"attr,permissions"`
	SAMLEnabled            bool                     `jsonapi:"attr,saml-enabled"`
	SessionRemember        int                      `jsonapi:"attr,session-remember"`
	SessionTimeout         int                      `jsonapi:"attr,session-timeout"`
	TrialExpiresAt         time.Time                `jsonapi:"attr,trial-expires-at,iso8601"`
	TwoFactorConformant    bool                     `jsonapi:"attr,two-factor-conformant"`
}

type OrganizationList struct {
	*Pagination
	Items []*Organization
}

// OrganizationListOptions represents the options for listing organizations.
type OrganizationListOptions struct {
	ListOptions
}

// OrganizationPermissions represents the organization permissions.
type OrganizationPermissions struct {
	CanCreateTeam               bool `json:"can-create-team"`
	CanCreateWorkspace          bool `json:"can-create-workspace"`
	CanCreateWorkspaceMigration bool `json:"can-create-workspace-migration"`
	CanDestroy                  bool `json:"can-destroy"`
	CanTraverse                 bool `json:"can-traverse"`
	CanUpdate                   bool `json:"can-update"`
	CanUpdateAPIToken           bool `json:"can-update-api-token"`
	CanUpdateOAuth              bool `json:"can-update-oauth"`
	CanUpdateSentinel           bool `json:"can-update-sentinel"`
}

func (org *Organization) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self": fmt.Sprintf("/api/v2/organizations/%s", org.Name),
	}
}

type OrganizationService interface {
	CreateOrganization(opts *tfe.OrganizationCreateOptions) (*Organization, error)
	GetOrganization(name string) (*Organization, error)
	ListOrganizations(opts OrganizationListOptions) (*OrganizationList, error)
	UpdateOrganization(name string, opts *tfe.OrganizationUpdateOptions) (*Organization, error)
	DeleteOrganization(name string) error
	GetEntitlements(name string) (*Entitlements, error)
}

func NewOrganizationFromOptions(opts *tfe.OrganizationCreateOptions, overrides ...func(*Organization)) (*Organization, error) {
	if !validString(opts.Name) {
		return nil, tfe.ErrRequiredName
	}
	if !validStringID(opts.Name) {
		return nil, tfe.ErrInvalidName
	}
	if !validString(opts.Email) {
		return nil, errors.New("email is required")
	}

	org := &Organization{
		Name:                   *opts.Name,
		Email:                  *opts.Email,
		Permissions:            &OrganizationPermissions{},
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

func UpdateOrganizationFromOptions(org *Organization, opts *tfe.OrganizationUpdateOptions) error {
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

// Entitlements represents the entitlements of an organization.
type Entitlements struct {
	ID                    string `jsonapi:"primary,entitlement-sets"`
	Agents                bool   `jsonapi:"attr,agents"`
	AuditLogging          bool   `jsonapi:"attr,audit-logging"`
	CostEstimation        bool   `jsonapi:"attr,cost-estimation"`
	Operations            bool   `jsonapi:"attr,operations"`
	PrivateModuleRegistry bool   `jsonapi:"attr,private-module-registry"`
	SSO                   bool   `jsonapi:"attr,sso"`
	Sentinel              bool   `jsonapi:"attr,sentinel"`
	StateStorage          bool   `jsonapi:"attr,state-storage"`
	Teams                 bool   `jsonapi:"attr,teams"`
	VCSIntegrations       bool   `jsonapi:"attr,vcs-integrations"`
}

func NewEntitlements(orgName string) *Entitlements {
	return &Entitlements{
		ID: orgName,
	}
}

func (e *Entitlements) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self": fmt.Sprintf("/api/v2/entitlement-set/%s", e.ID),
	}
}
