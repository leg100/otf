// Package rbac is concerned with authorization
package rbac

import "fmt"

var (
	// OrganizationMinPermissions are permissions granted to all team
	// members within an organization.
	OrganizationMinPermissions = Role{
		name: "minimum",
		permissions: map[Action]bool{
			GetOrganizationAction:  true,
			GetEntitlementsAction:  true,
			ListModulesAction:      true,
			GetModuleAction:        true,
			GetTeamAction:          true,
			ListTeamsAction:        true,
			GetUserAction:          true,
			ListUsersAction:        true,
			ListTagsAction:         true,
			ListVCSProvidersAction: true,
			GetVCSProviderAction:   true,
			ListVariableSetsAction: true,
			GetVariableSetAction:   true,
			WatchAgentsAction:      true,
			ListAgentsAction:       true,
		},
	}

	// WorkspaceReadRole is scoped to a workspace and permits read-only actions
	// on the workspace.
	WorkspaceReadRole = Role{
		name: "read",
		permissions: map[Action]bool{
			ListRunsAction:                       true,
			GetPlanFileAction:                    true,
			GetWorkspaceAction:                   true,
			GetStateVersionAction:                true,
			GetStateVersionOutputAction:          true,
			DownloadStateAction:                  true,
			DownloadConfigurationVersionAction:   true,
			GetRunAction:                         true,
			GetConfigurationVersionAction:        true,
			ListWorkspaceVariablesAction:         true,
			GetWorkspaceVariableAction:           true,
			WatchAction:                          true,
			ListWorkspaceTags:                    true,
			TailLogsAction:                       true,
			ListNotificationConfigurationsAction: true,
			GetNotificationConfigurationAction:   true,
		},
	}

	// WorkspacePlanRole is scoped to a workspace and permits creating plans on
	// the workspace.
	WorkspacePlanRole = Role{
		name: "plan",
		permissions: map[Action]bool{
			CreateRunAction:                  true,
			CreateConfigurationVersionAction: true,
		},
		inherits: &WorkspaceReadRole,
	}

	// WorkspaceWriteRole is scoped to a workspace and permits write actions on
	// the workspace.
	WorkspaceWriteRole = Role{
		name: "write",
		permissions: map[Action]bool{
			ApplyRunAction:                        true,
			CancelRunAction:                       true,
			LockWorkspaceAction:                   true,
			UnlockWorkspaceAction:                 true,
			CreateWorkspaceVariableAction:         true,
			UpdateWorkspaceVariableAction:         true,
			DeleteWorkspaceVariableAction:         true,
			CreateNotificationConfigurationAction: true,
			UpdateNotificationConfigurationAction: true,
			DeleteNotificationConfigurationAction: true,
		},
		inherits: &WorkspacePlanRole,
	}

	// WorkspaceAdminRole is scoped to a workspace and permits management of the
	// workspace.
	WorkspaceAdminRole = Role{
		name: "admin",
		permissions: map[Action]bool{
			SetWorkspacePermissionAction:   true,
			UnsetWorkspacePermissionAction: true,
			DeleteWorkspaceAction:          true,
			ForceUnlockWorkspaceAction:     true,
			UpdateWorkspaceAction:          true,
		},
		inherits: &WorkspaceWriteRole,
	}

	// WorkspaceManagerRole is scoped to an organization and permits management
	// of workspaces.
	WorkspaceManagerRole = Role{
		name: "workspace-manager",
		permissions: map[Action]bool{
			CreateWorkspaceAction: true,
			ListWorkspacesAction:  true,
			UpdateWorkspaceAction: true,
			AddTagsAction:         true,
			RemoveTagsAction:      true,
		},
		inherits: &WorkspaceAdminRole,
	}

	// VCSManagerRole is scoped to an organization and permits management of VCS
	// providers.
	VCSManagerRole = Role{
		name: "vcs-manager",
		permissions: map[Action]bool{
			CreateVCSProviderAction: true,
			DeleteVCSProviderAction: true,
		},
	}

	// RegistryManagerRole is scoped to an organization and permits management
	// of registry of modules and providers
	RegistryManagerRole = Role{
		name: "registry-manager",
		permissions: map[Action]bool{
			CreateModuleAction:        true,
			CreateModuleVersionAction: true,
			UpdateModuleAction:        true,
			DeleteModuleAction:        true,
		},
	}
)

// Role is a set of permitted actions
type Role struct {
	name        string
	permissions map[Action]bool
	inherits    *Role // inherit perms from this role too
}

func (r Role) IsAllowed(action Action) bool {
	if r.permissions[action] {
		return true
	}
	if r.inherits != nil {
		if r.inherits.IsAllowed(action) {
			return true
		}
	}
	return false
}

func (r Role) String() string {
	return r.name
}

func WorkspaceRoleFromString(role string) (Role, error) {
	switch role {
	case "read":
		return WorkspaceReadRole, nil
	case "plan":
		return WorkspacePlanRole, nil
	case "write":
		return WorkspaceWriteRole, nil
	case "admin":
		return WorkspaceAdminRole, nil
	default:
		return Role{}, fmt.Errorf("unknown role: %s", role)
	}
}
