package team

import (
	"github.com/leg100/otf/internal/resource"
)

type (
	// TFETeam represents an otf team.
	TFETeam struct {
		ID                 resource.TfeID         `jsonapi:"primary,teams"`
		Name               string                 `jsonapi:"attribute" json:"name"`
		OrganizationAccess *TFEOrganizationAccess `jsonapi:"attribute" json:"organization-access"`
		Visibility         string                 `jsonapi:"attribute" json:"visibility"`
		Permissions        *TFETeamPermissions    `jsonapi:"attribute" json:"permissions"`
		UserCount          int                    `jsonapi:"attribute" json:"users-count"`
		SSOTeamID          *string                `jsonapi:"attribute" json:"sso-team-id"`

		// Relations
		Users []*TFEUser `jsonapi:"relationship" json:"users"`
	}

	// TFEOrganizationAccess represents the team's permissions on its organization
	TFEOrganizationAccess struct {
		ManagePolicies        bool `jsonapi:"attribute" json:"manage-policies"`
		ManagePolicyOverrides bool `jsonapi:"attribute" json:"manage-policy-overrides"`
		ManageWorkspaces      bool `jsonapi:"attribute" json:"manage-workspaces"`
		ManageVCSSettings     bool `jsonapi:"attribute" json:"manage-vcs-settings"`
		ManageProviders       bool `jsonapi:"attribute" json:"manage-providers"`
		ManageModules         bool `jsonapi:"attribute" json:"manage-modules"`
		ManageRunTasks        bool `jsonapi:"attribute" json:"manage-run-tasks"`
		ManageProjects        bool `jsonapi:"attribute" json:"manage-projects"`
		ReadWorkspaces        bool `jsonapi:"attribute" json:"read-workspaces"`
		ReadProjects          bool `jsonapi:"attribute" json:"read-projects"`
		ManageMembership      bool `jsonapi:"attribute" json:"manage-membership"`
	}

	// TFETeamPermissions represents the current user's permissions on the team.
	TFETeamPermissions struct {
		CanDestroy          bool `jsonapi:"attribute" json:"can-destroy"`
		CanUpdateMembership bool `jsonapi:"attribute" json:"can-update-membership"`
	}

	// TFETeamCreateOptions represents the options for creating a team.
	TFETeamCreateOptions struct {
		// Type is a public field utilized by JSON:API to
		// set the resource type via the field tag.
		// It is not a user-defined value and does not need to be set.
		// https://jsonapi.org/format/#crud-creating
		Type string `jsonapi:"primary,teams"`

		// Name of the team.
		Name *string `jsonapi:"attribute" json:"name"`

		// Optional: Unique Identifier to control team membership via SAML
		SSOTeamID *string `jsonapi:"attribute" json:"sso-team-id,omitempty"`

		// The team's organization access
		OrganizationAccess *TFEOrganizationAccessOptions `jsonapi:"attribute" json:"organization-access,omitempty"`

		// The team's visibility ("secret", "organization")
		Visibility *string `jsonapi:"attribute" json:"visibility,omitempty"`
	}

	// TFETeamUpdateOptions represents the options for updating a team.
	TFETeamUpdateOptions struct {
		// Type is a public field utilized by JSON:API to
		// set the resource type via the field tag.
		// It is not a user-defined value and does not need to be set.
		// https://jsonapi.org/format/#crud-creating
		Type string `jsonapi:"primary,teams"`

		// Optional: New name for the team
		Name *string `jsonapi:"attribute" json:"name,omitempty"`

		// Optional: Unique Identifier to control team membership via SAML
		SSOTeamID *string `jsonapi:"attribute" json:"sso-team-id,omitempty"`

		// Optional: The team's organization access
		OrganizationAccess *TFEOrganizationAccessOptions `jsonapi:"attribute" json:"organization-access,omitempty"`

		// Optional: The team's visibility ("secret", "organization")
		Visibility *string `jsonapi:"attribute" json:"visibility,omitempty"`
	}

	// OrganizationAccessOptions represents the organization access options of a team.
	TFEOrganizationAccessOptions struct {
		ManagePolicies        *bool `json:"manage-policies,omitempty"`
		ManagePolicyOverrides *bool `json:"manage-policy-overrides,omitempty"`
		ManageWorkspaces      *bool `json:"manage-workspaces,omitempty"`
		ManageVCSSettings     *bool `json:"manage-vcs-settings,omitempty"`
		ManageProviders       *bool `json:"manage-providers,omitempty"`
		ManageModules         *bool `json:"manage-modules,omitempty"`
		ManageRunTasks        *bool `json:"manage-run-tasks,omitempty"`
		ManageProjects        *bool `json:"manage-projects,omitempty"`
		ReadWorkspaces        *bool `json:"read-workspaces,omitempty"`
		ReadProjects          *bool `json:"read-projects,omitempty"`
		ManageMembership      *bool `json:"manage-membership,omitempty"`
	}

	TFEUser struct {
		ID resource.TfeID `jsonapi:"primary,users"`
	}
)
