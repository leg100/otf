package dto

import (
	"time"
)

var (
	DefaultOrganizationPermissions = OrganizationPermissions{
		CanCreateWorkspace: true,
		CanUpdate:          true,
		CanDestroy:         true,
	}
)

// Organization JSON-API representation
type Organization struct {
	Name                  string                   `jsonapi:"primary,organizations"`
	CostEstimationEnabled bool                     `jsonapi:"attr,cost-estimation-enabled"`
	CreatedAt             time.Time                `jsonapi:"attr,created-at,iso8601"`
	ExternalID            string                   `jsonapi:"attr,external-id"`
	OwnersTeamSAMLRoleID  string                   `jsonapi:"attr,owners-team-saml-role-id"`
	Permissions           *OrganizationPermissions `jsonapi:"attr,permissions"`
	SAMLEnabled           bool                     `jsonapi:"attr,saml-enabled"`
	SessionRemember       int                      `jsonapi:"attr,session-remember"`
	SessionTimeout        int                      `jsonapi:"attr,session-timeout"`
	TrialExpiresAt        time.Time                `jsonapi:"attr,trial-expires-at,iso8601"`
	TwoFactorConformant   bool                     `jsonapi:"attr,two-factor-conformant"`
}

// OrganizationList JSON-API representation
type OrganizationList struct {
	*Pagination
	Items []*Organization
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

	SessionRemember *int `jsonapi:"attr,session-remember,omitempty"`

	// Session timeout after inactivity (minutes).
	SessionTimeout *int `jsonapi:"attr,session-timeout,omitempty"`
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

	// Session expiration (minutes).
	SessionRemember *int `jsonapi:"attr,session-remember,omitempty"`

	// Session timeout after inactivity (minutes).
	SessionTimeout *int `jsonapi:"attr,session-timeout,omitempty"`
}
