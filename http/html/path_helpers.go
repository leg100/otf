package html

import (
	"fmt"
	"html/template"
)

// organizationRoute provides info about a route for an organization resource
type organizationRoute interface {
	OrganizationName() string
}

// workspaceRoute provides info about a route for a workspace resource
type workspaceRoute interface {
	OrganizationName() string
	WorkspaceName() string
}

// runRoute provides info about a route for a run resource
type runRoute interface {
	// ID of run
	RunID() string
	// Name of run's workspace
	WorkspaceName() string
	// Name of run's organization
	OrganizationName() string
}

func loginPath() string {
	return "/login"
}

func logoutPath() string {
	return "/logout"
}

func getProfilePath() string {
	return "/profile"
}

func listSessionPath() string {
	return "/profile/sessions"
}

func revokeSessionPath() string {
	return "/profile/sessions/revoke"
}

func listTokenPath() string {
	return "/profile/tokens"
}

func deleteTokenPath() string {
	return "/profile/tokens/delete"
}

func newTokenPath() string {
	return "/profile/tokens/new"
}

func createTokenPath() string {
	return "/profile/tokens/create"
}

func listOrganizationPath() string {
	return "/organizations"
}

func newOrganizationPath() string {
	return "/organizations/new"
}

func createOrganizationPath() string {
	return "/organizations/create"
}

func getOrganizationPath(name organizationRoute) string {
	return fmt.Sprintf("/organizations/%s", name.OrganizationName())
}

func getOrganizationOverviewPath(name organizationRoute) string {
	return fmt.Sprintf("/organizations/%s/overview", name.OrganizationName())
}

func editOrganizationPath(name organizationRoute) string {
	return fmt.Sprintf("/organizations/%s/edit", name.OrganizationName())
}

func updateOrganizationPath(name organizationRoute) string {
	return fmt.Sprintf("/organizations/%s/update", name.OrganizationName())
}

func deleteOrganizationPath(name organizationRoute) string {
	return fmt.Sprintf("/organizations/%s/delete", name.OrganizationName())
}

func listWorkspacePath(name organizationRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces", name.OrganizationName())
}

func newWorkspacePath(name organizationRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/new", name.OrganizationName())
}

func createWorkspacePath(name organizationRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/create", name.OrganizationName())
}

func getWorkspacePath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s", ws.OrganizationName(), ws.WorkspaceName())
}

func editWorkspacePath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/edit", ws.OrganizationName(), ws.WorkspaceName())
}

func updateWorkspacePath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/update", ws.OrganizationName(), ws.WorkspaceName())
}

func deleteWorkspacePath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/delete", ws.OrganizationName(), ws.WorkspaceName())
}

func lockWorkspacePath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/lock", ws.OrganizationName(), ws.WorkspaceName())
}

func unlockWorkspacePath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/unlock", ws.OrganizationName(), ws.WorkspaceName())
}

func listRunPath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs", ws.OrganizationName(), ws.WorkspaceName())
}

func newRunPath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/new", ws.OrganizationName(), ws.WorkspaceName())
}

func createRunPath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/create", ws.OrganizationName(), ws.WorkspaceName())
}

func getRunPath(run runRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/%s", run.OrganizationName(), run.WorkspaceName(), run.RunID())
}

func getPlanPath(run runRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/%s/plan", run.OrganizationName(), run.WorkspaceName(), run.RunID())
}

func getApplyPath(run runRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/%s/apply", run.OrganizationName(), run.WorkspaceName(), run.RunID())
}

func deleteRunPath(run runRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/%s/delete", run.OrganizationName(), run.WorkspaceName(), run.RunID())
}

func addHelpersToFuncMap(m template.FuncMap) {
	m["loginPath"] = loginPath
	m["logoutPath"] = logoutPath
	m["getProfilePath"] = getProfilePath
	m["listSessionPath"] = listSessionPath
	m["revokeSessionPath"] = revokeSessionPath
	m["listTokenPath"] = listTokenPath
	m["deleteTokenPath"] = deleteTokenPath
	m["newTokenPath"] = newTokenPath
	m["createTokenPath"] = createTokenPath
	m["listOrganizationPath"] = listOrganizationPath
	m["newOrganizationPath"] = newOrganizationPath
	m["createOrganizationPath"] = createOrganizationPath
	m["getOrganizationPath"] = getOrganizationPath
	m["getOrganizationOverviewPath"] = getOrganizationOverviewPath
	m["editOrganizationPath"] = editOrganizationPath
	m["updateOrganizationPath"] = updateOrganizationPath
	m["deleteOrganizationPath"] = deleteOrganizationPath
	m["listWorkspacePath"] = listWorkspacePath
	m["newWorkspacePath"] = newWorkspacePath
	m["createWorkspacePath"] = createWorkspacePath
	m["getWorkspacePath"] = getWorkspacePath
	m["editWorkspacePath"] = editWorkspacePath
	m["updateWorkspacePath"] = updateWorkspacePath
	m["deleteWorkspacePath"] = deleteWorkspacePath
	m["lockWorkspacePath"] = lockWorkspacePath
	m["unlockWorkspacePath"] = unlockWorkspacePath
	m["listRunPath"] = listRunPath
	m["newRunPath"] = newRunPath
	m["createRunPath"] = createRunPath
	m["getRunPath"] = getRunPath
	m["getPlanPath"] = getPlanPath
	m["getApplyPath"] = getApplyPath
	m["deleteRunPath"] = deleteRunPath
}
