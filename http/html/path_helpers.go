package html

import (
	"fmt"
	"html/template"
	"path"
)

// organizationRoute provides info about a route for an organization resource
type organizationRoute interface {
	OrganizationName() string
}

// teamRoute provides info about a route for an team resource
type teamRoute interface {
	OrganizationName() string
	TeamName() string
}

// workspaceRoute provides info about a route for a workspace resource
type workspaceRoute interface {
	OrganizationName() string
	WorkspaceName() string
}

// vcsProviderRoute provides info about a route for a vcs provider resource
type vcsProviderRoute interface {
	OrganizationName() string
	WorkspaceName() string
	VCSProviderID() string
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

func adminLoginPath() string {
	return "/admin/login"
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

func listAgentTokenPath(route organizationRoute) string {
	return path.Join(getOrganizationPath(route), "agent-tokens")
}

func deleteAgentTokenPath(route organizationRoute) string {
	return path.Join(getOrganizationPath(route), "agent-tokens", "delete")
}

func createAgentTokenPath(route organizationRoute) string {
	return path.Join(getOrganizationPath(route), "agent-tokens", "create")
}

func newAgentTokenPath(route organizationRoute) string {
	return path.Join(getOrganizationPath(route), "agent-tokens", "new")
}

func listVCSProviderPath(route organizationRoute) string {
	return path.Join(getOrganizationPath(route), "vcs-providers")
}

func newVCSProviderPath(route organizationRoute, cloud string) string {
	return path.Join(getOrganizationPath(route), "vcs-providers", cloud, "new")
}

func createVCSProviderPath(route organizationRoute, cloud string) string {
	return path.Join(getOrganizationPath(route), "vcs-providers", cloud, "create")
}

func deleteVCSProviderPath(route organizationRoute) string {
	return path.Join(getOrganizationPath(route), "vcs-providers", "delete")
}

func listOrganizationPath() string {
	return "/organizations"
}

func getOrganizationPath(name organizationRoute) string {
	return fmt.Sprintf("/organizations/%s", name.OrganizationName())
}

func editOrganizationPath(name organizationRoute) string {
	return fmt.Sprintf("/organizations/%s/edit", name.OrganizationName())
}

func updateOrganizationPath(name organizationRoute) string {
	return fmt.Sprintf("/organizations/%s/update", name.OrganizationName())
}

func newOrganizationPath() string {
	return "/organizations/new"
}

func createOrganizationPath() string {
	return "/organizations/create"
}

func deleteOrganizationPath(name organizationRoute) string {
	return fmt.Sprintf("/organizations/%s/delete", name.OrganizationName())
}

func listUsersPath(name organizationRoute) string {
	return fmt.Sprintf("/organizations/%s/users", name.OrganizationName())
}

func listTeamsPath(name organizationRoute) string {
	return fmt.Sprintf("/organizations/%s/teams", name.OrganizationName())
}

func getTeamPath(name teamRoute) string {
	return fmt.Sprintf("/organizations/%s/teams/%s", name.OrganizationName(), name.TeamName())
}

func updateTeamPath(name teamRoute) string {
	return fmt.Sprintf("/organizations/%s/teams/%s/update", name.OrganizationName(), name.TeamName())
}

func listTeamUsersPath(name teamRoute) string {
	return fmt.Sprintf("/organizations/%s/teams/%s/users", name.OrganizationName(), name.TeamName())
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

func listWorkspaceVCSProvidersPath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/vcs-providers", ws.OrganizationName(), ws.WorkspaceName())
}

func listWorkspaceRepoPath(ws workspaceRoute, providerID string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/vcs-providers/%s/repos", ws.OrganizationName(), ws.WorkspaceName(), providerID)
}

func connectWorkspaceRepoPath(vcs vcsProviderRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/vcs-providers/%s/repos/connect", vcs.OrganizationName(), vcs.WorkspaceName(), vcs.VCSProviderID())
}

func disconnectWorkspaceRepoPath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/repo/disconnect", ws.OrganizationName(), ws.WorkspaceName())
}

func startRunPath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/start-run", ws.OrganizationName(), ws.WorkspaceName())
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

func setWorkspacePermissionPath(name workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/permissions", name.OrganizationName(), name.WorkspaceName())
}

func unsetWorkspacePermissionPath(name workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/permissions/unset", name.OrganizationName(), name.WorkspaceName())
}

func listRunPath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs", ws.OrganizationName(), ws.WorkspaceName())
}

func newRunPath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/new", ws.OrganizationName(), ws.WorkspaceName())
}

func getRunPath(run runRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/%s", run.OrganizationName(), run.WorkspaceName(), run.RunID())
}

func watchWorkspacePath(ws workspaceRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/watch", ws.OrganizationName(), ws.WorkspaceName())
}

func tailRunPath(run runRoute) string {
	return path.Join(getRunPath(run), "tail")
}

func deleteRunPath(run runRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/%s/delete", run.OrganizationName(), run.WorkspaceName(), run.RunID())
}

func cancelRunPath(run runRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/%s/cancel", run.OrganizationName(), run.WorkspaceName(), run.RunID())
}

func applyRunPath(run runRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/%s/apply", run.OrganizationName(), run.WorkspaceName(), run.RunID())
}

func discardRunPath(run runRoute) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/runs/%s/discard", run.OrganizationName(), run.WorkspaceName(), run.RunID())
}

func addHelpersToFuncMap(m template.FuncMap) {
	m["loginPath"] = loginPath
	m["logoutPath"] = logoutPath
	m["adminLoginPath"] = adminLoginPath
	m["getProfilePath"] = getProfilePath
	m["listSessionPath"] = listSessionPath
	m["revokeSessionPath"] = revokeSessionPath
	m["listTokenPath"] = listTokenPath
	m["deleteTokenPath"] = deleteTokenPath
	m["newTokenPath"] = newTokenPath
	m["createTokenPath"] = createTokenPath
	m["listOrganizationPath"] = listOrganizationPath
	m["getOrganizationPath"] = getOrganizationPath
	m["editOrganizationPath"] = editOrganizationPath
	m["updateOrganizationPath"] = updateOrganizationPath
	m["deleteOrganizationPath"] = deleteOrganizationPath
	m["newOrganizationPath"] = newOrganizationPath
	m["createOrganizationPath"] = createOrganizationPath
	m["listUsersPath"] = listUsersPath
	m["getTeamPath"] = getTeamPath
	m["updateTeamPath"] = updateTeamPath
	m["listTeamsPath"] = listTeamsPath
	m["listTeamUsersPath"] = listTeamUsersPath
	m["listWorkspacePath"] = listWorkspacePath
	m["newWorkspacePath"] = newWorkspacePath
	m["createWorkspacePath"] = createWorkspacePath
	m["getWorkspacePath"] = getWorkspacePath
	m["editWorkspacePath"] = editWorkspacePath
	m["updateWorkspacePath"] = updateWorkspacePath
	m["deleteWorkspacePath"] = deleteWorkspacePath
	m["lockWorkspacePath"] = lockWorkspacePath
	m["unlockWorkspacePath"] = unlockWorkspacePath
	m["setWorkspacePermissionPath"] = setWorkspacePermissionPath
	m["unsetWorkspacePermissionPath"] = unsetWorkspacePermissionPath
	m["listRunPath"] = listRunPath
	m["newRunPath"] = newRunPath
	m["getRunPath"] = getRunPath
	m["watchWorkspacePath"] = watchWorkspacePath
	m["tailRunPath"] = tailRunPath
	m["deleteRunPath"] = deleteRunPath
	m["cancelRunPath"] = cancelRunPath
	m["applyRunPath"] = applyRunPath
	m["discardRunPath"] = discardRunPath
	m["listAgentTokenPath"] = listAgentTokenPath
	m["deleteAgentTokenPath"] = deleteAgentTokenPath
	m["createAgentTokenPath"] = createAgentTokenPath
	m["newAgentTokenPath"] = newAgentTokenPath
	m["listVCSProviderPath"] = listVCSProviderPath
	m["newVCSProviderPath"] = newVCSProviderPath
	m["createVCSProviderPath"] = createVCSProviderPath
	m["deleteVCSProviderPath"] = deleteVCSProviderPath
	m["listWorkspaceVCSProvidersPath"] = listWorkspaceVCSProvidersPath
	m["listWorkspaceRepoPath"] = listWorkspaceRepoPath
	m["connectWorkspaceRepoPath"] = connectWorkspaceRepoPath
	m["disconnectWorkspaceRepoPath"] = disconnectWorkspaceRepoPath
	m["startRunPath"] = startRunPath
}
