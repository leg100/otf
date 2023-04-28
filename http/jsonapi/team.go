package jsonapi

type (
	// Team represents an otf team.
	Team struct {
		ID                 string              `jsonapi:"primary,teams"`
		Name               string              `jsonapi:"attr,name"`
		OrganizationAccess *OrganizationAccess `jsonapi:"attr,organization-access"`
		Visibility         string              `jsonapi:"attr,visibility"`
		Permissions        *TeamPermissions    `jsonapi:"attr,permissions"`
		UserCount          int                 `jsonapi:"attr,users-count"`
		SSOTeamID          string              `jsonapi:"attr,sso-team-id"`

		// Relations
		Users []*User `jsonapi:"relation,users"`
	}

	// OrganizationAccess represents the team's permissions on its organization
	OrganizationAccess struct {
		ManagePolicies        bool `jsonapi:"attr,manage-policies"`
		ManagePolicyOverrides bool `jsonapi:"attr,manage-policy-overrides"`
		ManageWorkspaces      bool `jsonapi:"attr,manage-workspaces"`
		ManageVCSSettings     bool `jsonapi:"attr,manage-vcs-settings"`
		ManageProviders       bool `jsonapi:"attr,manage-providers"`
		ManageModules         bool `jsonapi:"attr,manage-modules"`
		ManageRunTasks        bool `jsonapi:"attr,manage-run-tasks"`
		ManageProjects        bool `jsonapi:"attr,manage-projects"`
		ReadWorkspaces        bool `jsonapi:"attr,read-workspaces"`
		ReadProjects          bool `jsonapi:"attr,read-projects"`
		ManageMembership      bool `jsonapi:"attr,manage-membership"`
	}

	// TeamPermissions represents the current user's permissions on the team.
	TeamPermissions struct {
		CanDestroy          bool `jsonapi:"attr,can-destroy"`
		CanUpdateMembership bool `jsonapi:"attr,can-update-membership"`
	}

	// CreateTeamOptions represents the options for creating a
	// user.
	CreateTeamOptions struct {
		Type               string                     `jsonapi:"primary,teams"`
		Name               *string                    `jsonapi:"attr,name"`
		Organization       *string                    `schema:"organization_name,required"`
		OrganizationAccess *OrganizationAccessOptions `jsonapi:"attr,organization-access,omitempty"`
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
