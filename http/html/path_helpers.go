package html

import (
	"fmt"
	"html/template"
)

// organizationMeta provides organization name info
type organizationName interface {
	Name() string
}

// workspaceMeta provides workspace name info
type workspaceName interface {
	OrganizationName() string
	Name() string
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

func getOrganizationPath(name organizationName) string {
	return fmt.Sprintf("/organizations/%s", name.Name())
}

func getOrganizationOverviewPath(name organizationName) string {
	return fmt.Sprintf("/organizations/%s/overview", name.Name())
}

func editOrganizationPath(name organizationName) string {
	return fmt.Sprintf("/organizations/%s/edit", name.Name())
}

func updateOrganizationPath(name organizationName) string {
	return fmt.Sprintf("/organizations/%s/update", name.Name())
}

func deleteOrganizationPath(name organizationName) string {
	return fmt.Sprintf("/organizations/%s/delete", name.Name())
}

func listWorkspacePath(name organizationName) string {
	return fmt.Sprintf("/organizations/%s/workspaces", name.Name())
}

func newWorkspacePath(name organizationName) string {
	return fmt.Sprintf("/organizations/%s/workspaces/new", name.Name())
}

func createWorkspacePath(name organizationName) string {
	return fmt.Sprintf("/organizations/%s/workspaces/create", name.Name())
}

func getWorkspacePath(ws workspaceName) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s", ws.OrganizationName(), ws.Name())
}

func editWorkspacePath(ws workspaceName) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/edit", ws.OrganizationName(), ws.Name())
}

func updateWorkspacePath(ws workspaceName) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/update", ws.OrganizationName(), ws.Name())
}

func deleteWorkspacePath(ws workspaceName) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/delete", ws.OrganizationName(), ws.Name())
}

func lockWorkspacePath(ws workspaceName) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/lock", ws.OrganizationName(), ws.Name())
}

func unlockWorkspacePath(ws workspaceName) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/unlock", ws.OrganizationName(), ws.Name())
}

func listRunPath(ws workspaceName) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs", ws.OrganizationName(), ws.Name())
}

func newRunPath(ws workspaceName) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/new", ws.OrganizationName(), ws.Name())
}

func createRunPath(ws workspaceName) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/create", ws.OrganizationName(), ws.Name())
}

func getRunPath(organizationName, workspaceName, runID string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/%s", organizationName, workspaceName, runID)
}

func getPlanPath(organizationName, workspaceName, runID string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/%s/plan", organizationName, workspaceName, runID)
}

func getApplyPath(organizationName, workspaceName, runID string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/%s/apply", organizationName, workspaceName, runID)
}

func deleteRunPath(organizationName, workspaceName, runID string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/%s/delete", organizationName, workspaceName, runID)
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
