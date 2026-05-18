package authz

import (
	"fmt"

	"github.com/leg100/otf/internal/resource"
)

var (
	// OrganizationMinPermissions are permissions granted to all team
	// members within an organization.
	OrganizationMinPermissions = Role{
		name: "minimum",
		permissions: map[resource.Kind]map[resource.Action]bool{
			resource.OrganizationKind: map[resource.Action]bool{
				resource.Get: true,
			},
			resource.UserKind: map[resource.Action]bool{
				resource.Get:  true,
				resource.List: true,
			},
			resource.ModuleKind: map[resource.Action]bool{
				resource.Get:  true,
				resource.List: true,
			},
			resource.TeamKind: map[resource.Action]bool{
				resource.Get:  true,
				resource.List: true,
			},
			resource.VCSProviderKind: map[resource.Action]bool{
				resource.Get:  true,
				resource.List: true,
			},
			resource.VariableSetKind: map[resource.Action]bool{
				resource.Get:  true,
				resource.List: true,
			},
			resource.SSHKeyKind: map[resource.Action]bool{
				resource.Get:  true,
				resource.List: true,
			},
			resource.RunTriggerKind: map[resource.Action]bool{
				resource.Get:  true,
				resource.List: true,
			},
			resource.RunnerKind: map[resource.Action]bool{
				resource.Get:   true,
				resource.List:  true,
				resource.Watch: true,
			},
			resource.EntitlementKind: map[resource.Action]bool{
				resource.Get: true,
			},
			resource.TagKind: map[resource.Action]bool{
				resource.List: true,
			},
		},
	}

	// WorkspaceReadRole is scoped to a workspace and permits read-only actions
	// on the workspace.
	WorkspaceReadRole = Role{
		name: "read",
		permissions: map[resource.Kind]map[resource.Action]bool{
			resource.RunKind: map[resource.Action]bool{
				resource.Get:   true,
				resource.List:  true,
				resource.Watch: true,
			},
			resource.PlanFileKind: map[resource.Action]bool{
				resource.Get: true,
			},
			resource.WorkspaceKind: map[resource.Action]bool{
				resource.Get: true,
			},
			resource.StateVersionKind: map[resource.Action]bool{
				resource.Get:      true,
				resource.Download: true,
			},
			resource.StateVersionOutputKind: map[resource.Action]bool{
				resource.Get: true,
			},
			resource.ConfigVersionKind: map[resource.Action]bool{
				resource.Download: true,
				resource.Get:      true,
			},
			resource.NotificationConfigurationKind: map[resource.Action]bool{
				resource.Get:  true,
				resource.List: true,
			},
			resource.VariableKind: map[resource.Action]bool{
				resource.Get:  true,
				resource.List: true,
			},
			resource.ChunkKind: map[resource.Action]bool{
				resource.Tail: true,
			},
		},
	}

	// WorkspacePlanRole is scoped to a workspace and permits creating plans on
	// the workspace.
	WorkspacePlanRole = Role{
		name: "plan",
		permissions: map[resource.Kind]map[resource.Action]bool{
			resource.RunKind: map[resource.Action]bool{
				resource.Create: true,
			},
			resource.ConfigVersionKind: map[resource.Action]bool{
				resource.Create: true,
			},
		},
		inherits: &WorkspaceReadRole,
	}

	// WorkspaceWriteRole is scoped to a workspace and permits write actions on
	// the workspace.
	WorkspaceWriteRole = Role{
		name: "write",
		permissions: map[resource.Kind]map[resource.Action]bool{
			resource.RunKind: map[resource.Action]bool{
				resource.Apply:  true,
				resource.Cancel: true,
			},
			resource.WorkspaceKind: map[resource.Action]bool{
				resource.Lock:   true,
				resource.Unlock: true,
			},
			resource.NotificationConfigurationKind: map[resource.Action]bool{
				resource.Create: true,
				resource.Update: true,
				resource.Delete: true,
			},
			resource.VariableKind: map[resource.Action]bool{
				resource.Create: true,
				resource.Update: true,
				resource.Delete: true,
			},
		},
		inherits: &WorkspacePlanRole,
	}

	// WorkspaceAdminRole is scoped to a workspace and permits management of the
	// workspace.
	WorkspaceAdminRole = Role{
		name: "admin",
		permissions: map[resource.Kind]map[resource.Action]bool{
			resource.WorkspaceKind: map[resource.Action]bool{
				resource.Update:          true,
				resource.Delete:          true,
				resource.ForceUnlock:     true,
				resource.SetPermission:   true,
				resource.UnsetPermission: true,
			},
			resource.RunTriggerKind: map[resource.Action]bool{
				resource.Create: true,
				resource.Delete: true,
			},
		},
		inherits: &WorkspaceWriteRole,
	}

	// WorkspaceManagerRole is scoped to an organization and permits management
	// of workspaces.
	WorkspaceManagerRole = Role{
		name: "workspace-manager",
		permissions: map[resource.Kind]map[resource.Action]bool{
			resource.WorkspaceKind: map[resource.Action]bool{
				resource.Create: true,
				resource.List:   true,
				resource.Update: true,
			},
			resource.TagKind: map[resource.Action]bool{
				resource.Add:    true,
				resource.Remove: true,
			},
		},
		inherits: &WorkspaceAdminRole,
	}

	// VCSManagerRole is scoped to an organization and permits management of VCS
	// providers.
	VCSManagerRole = Role{
		name: "vcs-manager",
		permissions: map[resource.Kind]map[resource.Action]bool{
			resource.VCSProviderKind: map[resource.Action]bool{
				resource.Create: true,
				resource.Delete: true,
			},
		},
	}

	// RegistryManagerRole is scoped to an organization and permits management
	// of registry of modules and providers
	RegistryManagerRole = Role{
		name: "registry-manager",
		permissions: map[resource.Kind]map[resource.Action]bool{
			resource.ModuleKind: map[resource.Action]bool{
				resource.Create: true,
				resource.Update: true,
				resource.Delete: true,
			},
			resource.ModuleVersionKind: map[resource.Action]bool{
				resource.Create: true,
			},
		},
	}
)

// Role is a set of permitted actions
type Role struct {
	name        string
	permissions map[resource.Kind]map[resource.Action]bool
	inherits    *Role // inherit perms from this role too
}

func (r Role) IsAllowed(action resource.Action, kind resource.Kind) bool {
	if r.permissions[kind][action] {
		return true
	}
	if r.inherits != nil {
		if r.inherits.IsAllowed(action, kind) {
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
