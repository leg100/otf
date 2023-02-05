// Package rbac is concerned with authorization
package rbac

import "fmt"

var (
	// OrganizationGuestRole is scoped to an organization and permits
	// lowly-privileged actions to all user members.
	OrganizationGuestRole = Role{
		name: "registry-manager",
		permissions: map[Action]bool{
			GetOrganizationAction: true, // guest can read org info
			GetEntitlementsAction: true, // guest can read entitlements
			ListModulesAction:     true, // guest can list mods within org
			GetModuleAction:       true, // guest can read mod info
		},
	}

	// WorkspaceReadRole is scoped to a workspace and permits read-only actions
	// on the workspace.
	WorkspaceReadRole = Role{
		name: "read",
		permissions: map[Action]bool{
			ListRunsAction:                     true,
			GetPlanFileAction:                  true,
			GetWorkspaceAction:                 true,
			GetStateVersionAction:              true,
			DownloadStateAction:                true,
			DownloadConfigurationVersionAction: true,
			GetRunAction:                       true,
			GetConfigurationVersionAction:      true,
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
	}

	// WorkspaceWriteRole is scoped to a workspace and permits write actions on
	// the workspace.
	WorkspaceWriteRole = Role{
		name: "write",
		permissions: map[Action]bool{
			ApplyRunAction:        true,
			LockWorkspaceAction:   true,
			UnlockWorkspaceAction: true,
		},
	}

	// WorkspaceAdminRole is scoped to a workspace and permits management of the
	// workspace.
	WorkspaceAdminRole = Role{
		name: "admin",
		permissions: map[Action]bool{
			GetConfigurationVersionAction: true,
			SetWorkspacePermissionAction:  true,
			DeleteWorkspaceAction:         true,
		},
	}

	// WorkspaceManagerRole is scoped to an organization and permits management
	// of workspaces.
	WorkspaceManagerRole = Role{
		name: "workspace-manager",
		permissions: map[Action]bool{
			CreateWorkspaceAction:          true,
			ListWorkspacesAction:           true,
			UpdateWorkspaceAction:          true,
			SetWorkspacePermissionAction:   true,
			UnsetWorkspacePermissionAction: true,
		},
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

func init() {
	// plan role includes read permissions
	for p := range WorkspaceReadRole.permissions {
		WorkspacePlanRole.permissions[p] = true
	}
	// write role includes plan permissions
	for p := range WorkspacePlanRole.permissions {
		WorkspaceWriteRole.permissions[p] = true
	}
	// admin role includes write permissions
	for p := range WorkspaceWriteRole.permissions {
		WorkspaceAdminRole.permissions[p] = true
	}
	// workspace manager role includes admin permissions
	for p := range WorkspaceAdminRole.permissions {
		WorkspaceManagerRole.permissions[p] = true
	}
}

// Role is a set of permitted actions
type Role struct {
	name        string
	permissions map[Action]bool
}

func (r Role) IsAllowed(action Action) bool {
	return r.permissions[action]
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
