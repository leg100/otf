package organization

import (
	"time"

	"github.com/leg100/otf/internal/resource"
)

var DefaultOrganizationPermissions = TFEOrganizationPermissions{
	CanCreateWorkspace: true,
	CanUpdate:          true,
	CanDestroy:         true,
}

// TFEOrganization represents a Terraform Enterprise organization.
type TFEOrganization struct {
	Name                                              Name                        `jsonapi:"primary,organizations"`
	AssessmentsEnforced                               bool                        `jsonapi:"attribute" json:"assessments-enforced"`
	CollaboratorAuthPolicy                            TFEAuthPolicyType           `jsonapi:"attribute" json:"collaborator-auth-policy"`
	CostEstimationEnabled                             bool                        `jsonapi:"attribute" json:"cost-estimation-enabled"`
	CreatedAt                                         time.Time                   `jsonapi:"attribute" json:"created-at"`
	Email                                             string                      `jsonapi:"attribute" json:"email"`
	ExternalID                                        resource.TfeID              `jsonapi:"attribute" json:"external-id"`
	OwnersTeamSAMLRoleID                              resource.TfeID              `jsonapi:"attribute" json:"owners-team-saml-role-id"`
	Permissions                                       *TFEOrganizationPermissions `jsonapi:"attribute" json:"permissions"`
	SAMLEnabled                                       bool                        `jsonapi:"attribute" json:"saml-enabled"`
	SessionRemember                                   *int                        `jsonapi:"attribute" json:"session-remember"`
	SessionTimeout                                    *int                        `jsonapi:"attribute" json:"session-timeout"`
	TrialExpiresAt                                    time.Time                   `jsonapi:"attribute" json:"trial-expires-at"`
	TwoFactorConformant                               bool                        `jsonapi:"attribute" json:"two-factor-conformant"`
	SendPassingStatusesForUntriggeredSpeculativePlans bool                        `jsonapi:"attribute" json:"send-passing-statuses-for-untriggered-speculative-plans"`
	RemainingTestableCount                            int                         `jsonapi:"attribute" json:"remaining-testable-count"`

	// Note: This will be false for TFE versions older than v202211, where the setting was introduced.
	// On those TFE versions, safe delete does not exist, so ALL deletes will be force deletes.
	AllowForceDeleteWorkspaces bool `jsonapi:"attribute" json:"allow-force-delete-workspaces"`

	// Relations
	// DefaultProject *Project `jsonapi:"relation,default-project"`
}

// TFEOrganizationPermissions represents the organization permissions.
type TFEOrganizationPermissions struct {
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

// TFEOrganizationCreateOptions represents the options for creating an organization.
type TFEOrganizationCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,organizations"`

	// Required: Name of the organization.
	Name *string `jsonapi:"attribute" json:"name"`

	// Optional: AssessmentsEnforced toggles whether health assessment enablement is enforced across all assessable workspaces (those with a minimum terraform versio of 0.15.4 and not running in local execution mode) or if the decision to enabled health assessments is delegated to the workspace setting AssessmentsEnabled.
	AssessmentsEnforced *bool `jsonapi:"attribute" json:"assessments-enforced,omitempty"`

	// Required: Admin email address.
	Email *string `jsonapi:"attribute" json:"email"`

	// Optional: Session expiration (minutes).
	SessionRemember *int `jsonapi:"attribute" json:"session-remember,omitempty"`

	// Optional: Session timeout after inactivity (minutes).
	SessionTimeout *int `jsonapi:"attribute" json:"session-timeout,omitempty"`

	// Optional: Authentication policy.
	CollaboratorAuthPolicy *TFEAuthPolicyType `jsonapi:"attribute" json:"collaborator-auth-policy,omitempty"`

	// Optional: Enable Cost Estimation
	CostEstimationEnabled *bool `jsonapi:"attribute" json:"cost-estimation-enabled,omitempty"`

	// Optional: The name of the "owners" team
	OwnersTeamSAMLRoleID *string `jsonapi:"attribute" json:"owners-team-saml-role-id,omitempty"`

	// Optional: SendPassingStatusesForUntriggeredSpeculativePlans toggles behavior of untriggered speculative plans to send status updates to version control systems like GitHub.
	SendPassingStatusesForUntriggeredSpeculativePlans *bool `jsonapi:"attribute" json:"send-passing-statuses-for-untriggered-speculative-plans,omitempty"`

	// Optional: AllowForceDeleteWorkspaces toggles behavior of allowing workspace admins to delete workspaces with resources under management.
	AllowForceDeleteWorkspaces *bool `jsonapi:"attribute" json:"allow-force-delete-workspaces,omitempty"`
}

// TFEOrganizationUpdateOptions represents the options for updating an organization.
type TFEOrganizationUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,organizations"`

	// New name for the organization.
	Name *string `jsonapi:"attribute" json:"name,omitempty"`

	// Optional: AssessmentsEnforced toggles whether health assessment enablement is enforced across all assessable workspaces (those with a minimum terraform versio of 0.15.4 and not running in local execution mode) or if the decision to enabled health assessments is delegated to the workspace setting AssessmentsEnabled.
	AssessmentsEnforced *bool `jsonapi:"attribute" json:"assessments-enforced,omitempty"`

	// New admin email address.
	Email *string `jsonapi:"attribute" json:"email,omitempty"`

	// Session expiration (minutes).
	SessionRemember *int `jsonapi:"attribute" json:"session-remember,omitempty"`

	// Session timeout after inactivity (minutes).
	SessionTimeout *int `jsonapi:"attribute" json:"session-timeout,omitempty"`

	// Authentication policy.
	CollaboratorAuthPolicy *TFEAuthPolicyType `jsonapi:"attribute" json:"collaborator-auth-policy,omitempty"`

	// Enable Cost Estimation
	CostEstimationEnabled *bool `jsonapi:"attribute" json:"cost-estimation-enabled,omitempty"`

	// The name of the "owners" team
	OwnersTeamSAMLRoleID *string `jsonapi:"attribute" json:"owners-team-saml-role-id,omitempty"`

	// SendPassingStatusesForUntriggeredSpeculativePlans toggles behavior of untriggered speculative plans to send status updates to version control systems like GitHub.
	SendPassingStatusesForUntriggeredSpeculativePlans *bool `jsonapi:"attribute" json:"send-passing-statuses-for-untriggered-speculative-plans,omitempty"`

	// Optional: AllowForceDeleteWorkspaces toggles behavior of allowing workspace admins to delete workspaces with resources under management.
	AllowForceDeleteWorkspaces *bool `jsonapi:"attribute" json:"allow-force-delete-workspaces,omitempty"`
}

// Entitlements represents the entitlements of an organization. Unlike TFE/TFC,
// OTF is free and therefore the user is entitled to all currently supported
// services.  Entitlements represents the entitlements of an organization.
type TFEEntitlements struct {
	ID                    resource.TfeID `jsonapi:"primary,entitlement-sets"`
	Agents                bool           `jsonapi:"attribute" json:"agents"`
	AuditLogging          bool           `jsonapi:"attribute" json:"audit-logging"`
	CostEstimation        bool           `jsonapi:"attribute" json:"cost-estimation"`
	Operations            bool           `jsonapi:"attribute" json:"operations"`
	PrivateModuleRegistry bool           `jsonapi:"attribute" json:"private-module-registry"`
	SSO                   bool           `jsonapi:"attribute" json:"sso"`
	Sentinel              bool           `jsonapi:"attribute" json:"sentinel"`
	StateStorage          bool           `jsonapi:"attribute" json:"state-storage"`
	Teams                 bool           `jsonapi:"attribute" json:"teams"`
	VCSIntegrations       bool           `jsonapi:"attribute" json:"vcs-integrations"`
}

// TFEAuthPolicyType represents an authentication policy type.
type TFEAuthPolicyType string

// List of available authentication policies.
const (
	AuthPolicyPassword  TFEAuthPolicyType = "password"
	AuthPolicyTwoFactor TFEAuthPolicyType = "two_factor_mandatory"
)

// TFEOrganizationToken represents a Terraform Enterprise organization token.
type TFEOrganizationToken struct {
	ID        resource.TfeID `jsonapi:"primary,authentication-tokens"`
	CreatedAt time.Time      `jsonapi:"attribute" json:"created-at"`
	Token     string         `jsonapi:"attribute" json:"token"`
	ExpiredAt *time.Time     `jsonapi:"attribute" json:"expired-at"`
}

// TFEOrganizationTokenCreateOptions contains the options for creating an organization token.
type TFEOrganizationTokenCreateOptions struct {
	// Optional: The token's expiration date.
	// This feature is available in TFE release v202305-1 and later
	ExpiredAt *time.Time `jsonapi:"attribute" json:"expired-at,omitempty"`
}
