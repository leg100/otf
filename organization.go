package otf

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	DefaultSessionTimeout         = 20160
	DefaultSessionExpiration      = 20160
	DefaultCollaboratorAuthPolicy = "password"
	DefaultCostEstimationEnabled  = true

	// List of available authentication policies.
	AuthPolicyPassword  AuthPolicyType = "password"
	AuthPolicyTwoFactor AuthPolicyType = "two_factor_mandatory"
)

var (
	DefaultOrganizationPermissions = OrganizationPermissions{
		CanCreateWorkspace: true,
		CanUpdate:          true,
		CanDestroy:         true,
	}
)

// AuthPolicyType represents an authentication policy type.
type AuthPolicyType string

// Organization represents a Terraform Enterprise organization.
type Organization struct {
	ID string

	gorm.Model

	Name                   string
	CollaboratorAuthPolicy AuthPolicyType
	CostEstimationEnabled  bool
	Email                  string
	OwnersTeamSAMLRoleID   string
	Permissions            *OrganizationPermissions
	SAMLEnabled            bool
	SessionRemember        int
	SessionTimeout         int
	TrialExpiresAt         time.Time
	TwoFactorConformant    bool
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

// OrganizationCreateOptions represents the options for creating an
// organization.
type OrganizationCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,organizations"`

	// Name of the organization.
	Name *string `jsonapi:"attr,name"`

	// Admin email address.
	Email *string `jsonapi:"attr,email"`

	// Session expiration (minutes).
	SessionRemember *int `jsonapi:"attr,session-remember,omitempty"`

	// Session timeout after inactivity (minutes).
	SessionTimeout *int `jsonapi:"attr,session-timeout,omitempty"`

	// Authentication policy.
	CollaboratorAuthPolicy *AuthPolicyType `jsonapi:"attr,collaborator-auth-policy,omitempty"`

	// Enable Cost Estimation
	CostEstimationEnabled *bool `jsonapi:"attr,cost-estimation-enabled,omitempty"`

	// The name of the "owners" team
	OwnersTeamSAMLRoleID *string `jsonapi:"attr,owners-team-saml-role-id,omitempty"`
}

// OrganizationUpdateOptions represents the options for updating an
// organization.
type OrganizationUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,organizations"`

	// New name for the organization.
	Name *string `jsonapi:"attr,name,omitempty"`

	// New admin email address.
	Email *string `jsonapi:"attr,email,omitempty"`

	// Session expiration (minutes).
	SessionRemember *int `jsonapi:"attr,session-remember,omitempty"`

	// Session timeout after inactivity (minutes).
	SessionTimeout *int `jsonapi:"attr,session-timeout,omitempty"`

	// Authentication policy.
	CollaboratorAuthPolicy *AuthPolicyType `jsonapi:"attr,collaborator-auth-policy,omitempty"`

	// Enable Cost Estimation
	CostEstimationEnabled *bool `jsonapi:"attr,cost-estimation-enabled,omitempty"`

	// The name of the "owners" team
	OwnersTeamSAMLRoleID *string `jsonapi:"attr,owners-team-saml-role-id,omitempty"`
}

// OrganizationList represents a list of Organizations.
type OrganizationList struct {
	*Pagination
	Items []*Organization
}

// OrganizationListOptions represents the options for listing organizations.
type OrganizationListOptions struct {
	ListOptions
}

type OrganizationService interface {
	Create(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error)
	Get(name string) (*Organization, error)
	List(opts OrganizationListOptions) (*OrganizationList, error)
	Update(name string, opts *OrganizationUpdateOptions) (*Organization, error)
	Delete(name string) error
	GetEntitlements(name string) (*Entitlements, error)
}

type OrganizationStore interface {
	Create(org *Organization) (*Organization, error)
	Get(name string) (*Organization, error)
	List(opts OrganizationListOptions) (*OrganizationList, error)
	Update(name string, fn func(*Organization) error) (*Organization, error)
	Delete(name string) error
}

func (o OrganizationCreateOptions) Valid() error {
	if !validString(o.Name) {
		return ErrRequiredName
	}
	if !ValidStringID(o.Name) {
		return ErrInvalidName
	}
	if !validString(o.Email) {
		return errors.New("email is required")
	}
	return nil
}

func NewOrganization(opts OrganizationCreateOptions) (*Organization, error) {
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

func UpdateOrganization(org *Organization, opts *OrganizationUpdateOptions) error {
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

	return nil
}
