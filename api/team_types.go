package api

type (
	// Team represents an otf team.
	Team struct {
		ID                 string              `jsonapi:"primary,teams"`
		Name               string              `jsonapi:"attribute" json:"name"`
		OrganizationAccess *OrganizationAccess `jsonapi:"attribute" json:"organization-access"`
		Visibility         string              `jsonapi:"attribute" json:"visibility"`
		Permissions        *TeamPermissions    `jsonapi:"attribute" json:"permissions"`
		UserCount          int                 `jsonapi:"attribute" json:"users-count"`
		SSOTeamID          string              `jsonapi:"attribute" json:"sso-team-id"`

		// Relations
		Users []*User `jsonapi:"relationship" json:"users"`
	}

	// OrganizationAccess represents the team's permissions on its organization
	OrganizationAccess struct {
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

	// TeamPermissions represents the current user's permissions on the team.
	TeamPermissions struct {
		CanDestroy          bool `jsonapi:"attribute" json:"can-destroy"`
		CanUpdateMembership bool `jsonapi:"attribute" json:"can-update-membership"`
	}

	// CreateTeamOptions represents the options for creating a
	// user.
	CreateTeamOptions struct {
		Type               string                     `jsonapi:"primary,teams"`
		Name               *string                    `jsonapi:"attribute" json:"name"`
		Organization       *string                    `schema:"organization_name,required"`
		OrganizationAccess *OrganizationAccessOptions `jsonapi:"attribute" json:"organization-access,omitempty"`
	}

	// OrganizationAccessOptions represents the organization access options of a team.
	OrganizationAccessOptions struct {
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
)
