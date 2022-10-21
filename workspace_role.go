package otf

import "fmt"

var (
	WorkspaceReadRole = WorkspaceRole{
		name: "read",
		permissions: map[Action]bool{
			ListRunsAction:                true,
			GetPlanFileAction:             true,
			GetWorkspaceAction:            true,
			GetStateVersionAction:         true,
			GetRunAction:                  true,
			GetConfigurationVersionAction: true,
		},
	}

	WorkspacePlanRole = WorkspaceRole{
		name: "plan",
		permissions: map[Action]bool{
			CreateRunAction:                  true,
			CreateConfigurationVersionAction: true,
		},
	}

	WorkspaceWriteRole = WorkspaceRole{
		name: "write",
		permissions: map[Action]bool{
			ApplyRunAction:        true,
			LockWorkspaceAction:   true,
			UnlockWorkspaceAction: true,
		},
	}

	WorkspaceAdminRole = WorkspaceRole{
		name: "admin",
		permissions: map[Action]bool{
			GetConfigurationVersionAction: true,
			SetWorkspacePermissionAction:  true,
			DeleteWorkspaceAction:         true,
		},
	}

	WorkspaceManagerRole = WorkspaceRole{
		name: "workspace-manager",
		permissions: map[Action]bool{
			CreateWorkspaceAction:          true,
			ListWorkspacesAction:           true,
			UpdateWorkspaceAction:          true,
			SetWorkspacePermissionAction:   true,
			UnsetWorkspacePermissionAction: true,
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

// WorkspaceRole is a set of permitted actions
type WorkspaceRole struct {
	name        string
	permissions map[Action]bool
}

func (r WorkspaceRole) IsAllowed(action Action) bool {
	return r.permissions[action]
}

func (r WorkspaceRole) String() string {
	return r.name
}

func WorkspaceRoleFromString(role string) (WorkspaceRole, error) {
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
		return WorkspaceRole{}, fmt.Errorf("unknown role: %s", role)
	}
}
