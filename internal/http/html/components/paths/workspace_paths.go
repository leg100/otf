// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/a-h/templ"
)

func Workspaces(organization string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/workspaces", organization))
}

func CreateWorkspace(organization string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/workspaces/create", organization))
}

func NewWorkspace(organization string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/workspaces/new", organization))
}

func Workspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s", workspace))
}

func EditWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/edit", workspace))
}

func UpdateWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/update", workspace))
}

func DeleteWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/delete", workspace))
}

func LockWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/lock", workspace))
}

func UnlockWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/unlock", workspace))
}

func ForceUnlockWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/force-unlock", workspace))
}

func SetPermissionWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/set-permission", workspace))
}

func UnsetPermissionWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/unset-permission", workspace))
}

func WatchWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/watch", workspace))
}

func ConnectWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/connect", workspace))
}

func DisconnectWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/disconnect", workspace))
}

func StartRunWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/start-run", workspace))
}

func SetupConnectionProviderWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/setup-connection-provider", workspace))
}

func SetupConnectionRepoWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/setup-connection-repo", workspace))
}

func CreateTagWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/create-tag", workspace))
}

func DeleteTagWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/delete-tag", workspace))
}

func StateWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/state", workspace))
}

func PoolsWorkspace(workspace string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/pools", workspace))
}
