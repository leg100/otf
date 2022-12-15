package html

import (
	"fmt"
	"html/template"
	"path"

	"github.com/leg100/otf"
)

func loginPath() string {
	return "/login"
}

func logoutPath() string {
	return "/logout"
}

func adminLoginPath() string {
	return "/admin/login"
}

func profilePath() string {
	return "/profile"
}

func sessionsPath() string {
	return "/profile/sessions"
}

func revokeSessionPath() string {
	return "/profile/sessions/revoke"
}

func tokensPath() string {
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

func agentTokensPath(organization string) string {
	return path.Join("/organizations", organization, "agent-tokens")
}

func deleteAgentTokenPath(organization string) string {
	return path.Join("/organizations", organization, "agent-tokens", "delete")
}

func createAgentTokenPath(organization string) string {
	return path.Join("/organizations", organization, "agent-tokens", "create")
}

func newAgentTokenPath(organization string) string {
	return path.Join("/organizations", organization, "agent-tokens", "new")
}

func vcsProvidersPath(organization string) string {
	return path.Join("/organizations", organization, "vcs-providers")
}

func newVCSProviderPath(organization string, cloud string) string {
	return path.Join("/organizations", organization, "vcs-providers", cloud, "new")
}

func createVCSProviderPath(organization string, cloud string) string {
	return path.Join("/organizations", organization, "vcs-providers", cloud, "create")
}

func deleteVCSProviderPath(organization string) string {
	return path.Join("/organizations", organization, "vcs-providers", "delete")
}

func organizationsPath() string {
	return "/organizations"
}

func organizationPath(organization string) string {
	return fmt.Sprintf("/organizations/%s", organization)
}

func editOrganizationPath(organization string) string {
	return fmt.Sprintf("/organizations/%s/edit", organization)
}

func updateOrganizationPath(organization string) string {
	return fmt.Sprintf("/organizations/%s/update", organization)
}

func newOrganizationPath() string {
	return "/organizations/new"
}

func createOrganizationPath() string {
	return "/organizations/create"
}

func deleteOrganizationPath(organization string) string {
	return fmt.Sprintf("/organizations/%s/delete", organization)
}

func usersPath(organization string) string {
	return fmt.Sprintf("/organizations/%s/users", organization)
}

func teamsPath(organization string) string {
	return fmt.Sprintf("/organizations/%s/teams", organization)
}

func teamPath(teamID string) string {
	return fmt.Sprintf("/teams/%s", teamID)
}

func updateTeamPath(teamID string) string {
	return fmt.Sprintf("/teams/%s/update", teamID)
}

func teamUsersPath(teamID string) string {
	return fmt.Sprintf("/teams/%s/users", teamID)
}

func workspacesPath(organization string) string {
	return fmt.Sprintf("/organizations/%s/workspaces", organization)
}

func newWorkspacePath(organization string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/new", organization)
}

func createWorkspacePath(organization string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/create", organization)
}

func workspacePath(workspaceID string) string {
	return fmt.Sprintf("/workspaces/%s", workspaceID)
}

func workspaceVCSProvidersPath(workspaceID string) string {
	return fmt.Sprintf("/workspaces/%s/vcs-providers", workspaceID)
}

func workspaceReposPath(workspaceID, providerID string) string {
	return fmt.Sprintf("/workspaces/%s/vcs-providers/%s/repos", workspaceID, providerID)
}

func connectWorkspaceRepoPath(workspaceID, providerID string) string {
	return fmt.Sprintf("/workspaces/%s/vcs-providers/%s/repos/connect", workspaceID, providerID)
}

func disconnectWorkspaceRepoPath(workspaceID string) string {
	return fmt.Sprintf("/workspaces/%s/repo/disconnect", workspaceID)
}

func startRunPath(workspaceID string) string {
	return fmt.Sprintf("/workspaces/%s/start-run", workspaceID)
}

func editWorkspacePath(workspaceID string) string {
	return fmt.Sprintf("/workspaces/%s/edit", workspaceID)
}

func updateWorkspacePath(workspaceID string) string {
	return fmt.Sprintf("/workspaces/%s/update", workspaceID)
}

func deleteWorkspacePath(workspaceID string) string {
	return fmt.Sprintf("/workspaces/%s/delete", workspaceID)
}

func lockWorkspacePath(workspaceID string) string {
	return fmt.Sprintf("/workspaces/%s/lock", workspaceID)
}

func unlockWorkspacePath(workspaceID string) string {
	return fmt.Sprintf("/workspaces/%s/unlock", workspaceID)
}

func setWorkspacePermissionPath(workspaceID string) string {
	return fmt.Sprintf("/workspaces/%s/permissions", workspaceID)
}

func unsetWorkspacePermissionPath(workspaceID string) string {
	return fmt.Sprintf("/workspaces/%s/permissions/unset", workspaceID)
}

func workspaceRunsPath(workspaceID string) string {
	return fmt.Sprintf("/workspaces/%s/runs", workspaceID)
}

func newRunPath(workspaceID string) string {
	return fmt.Sprintf("/workspaces/%s/runs/new", workspaceID)
}

func runPath(runID string) string {
	return otf.RunGetPathUI(runID)
}

func watchWorkspacePath(workspaceID string) string {
	return fmt.Sprintf("/workspaces/%s/watch", workspaceID)
}

func tailRunPath(runID string) string {
	return path.Join("/runs", runID, "tail")
}

func deleteRunPath(runID string) string {
	return fmt.Sprintf("/runs/%s/delete", runID)
}

func cancelRunPath(runID string) string {
	return fmt.Sprintf("runs/%s/cancel", runID)
}

func applyRunPath(runID string) string {
	return fmt.Sprintf("/runs/%s/apply", runID)
}

func discardRunPath(runID string) string {
	return fmt.Sprintf("/runs/%s/discard", runID)
}

func addHelpersToFuncMap(m template.FuncMap) {
	m["loginPath"] = loginPath
	m["logoutPath"] = logoutPath
	m["adminLoginPath"] = adminLoginPath
	m["profilePath"] = profilePath
	m["sessionsPath"] = sessionsPath
	m["revokeSessionPath"] = revokeSessionPath
	m["tokensPath"] = tokensPath
	m["deleteTokenPath"] = deleteTokenPath
	m["newTokenPath"] = newTokenPath
	m["createTokenPath"] = createTokenPath
	m["organizationsPath"] = organizationsPath
	m["organizationPath"] = organizationPath
	m["editOrganizationPath"] = editOrganizationPath
	m["updateOrganizationPath"] = updateOrganizationPath
	m["deleteOrganizationPath"] = deleteOrganizationPath
	m["newOrganizationPath"] = newOrganizationPath
	m["createOrganizationPath"] = createOrganizationPath
	m["usersPath"] = usersPath
	m["teamPath"] = teamPath
	m["updateTeamPath"] = updateTeamPath
	m["teamsPath"] = teamsPath
	m["teamUsersPath"] = teamUsersPath
	m["workspacesPath"] = workspacesPath
	m["newWorkspacePath"] = newWorkspacePath
	m["createWorkspacePath"] = createWorkspacePath
	m["workspacePath"] = workspacePath
	m["editWorkspacePath"] = editWorkspacePath
	m["updateWorkspacePath"] = updateWorkspacePath
	m["deleteWorkspacePath"] = deleteWorkspacePath
	m["lockWorkspacePath"] = lockWorkspacePath
	m["unlockWorkspacePath"] = unlockWorkspacePath
	m["setWorkspacePermissionPath"] = setWorkspacePermissionPath
	m["unsetWorkspacePermissionPath"] = unsetWorkspacePermissionPath
	m["workspaceRunsPath"] = workspaceRunsPath
	m["newRunPath"] = newRunPath
	m["runPath"] = runPath
	m["watchWorkspacePath"] = watchWorkspacePath
	m["tailRunPath"] = tailRunPath
	m["deleteRunPath"] = deleteRunPath
	m["cancelRunPath"] = cancelRunPath
	m["applyRunPath"] = applyRunPath
	m["discardRunPath"] = discardRunPath
	m["agentTokensPath"] = agentTokensPath
	m["deleteAgentTokenPath"] = deleteAgentTokenPath
	m["createAgentTokenPath"] = createAgentTokenPath
	m["newAgentTokenPath"] = newAgentTokenPath
	m["vcsProvidersPath"] = vcsProvidersPath
	m["newVCSProviderPath"] = newVCSProviderPath
	m["createVCSProviderPath"] = createVCSProviderPath
	m["deleteVCSProviderPath"] = deleteVCSProviderPath
	m["workspaceVCSProvidersPath"] = workspaceVCSProvidersPath
	m["workspaceReposPath"] = workspaceReposPath
	m["connectWorkspaceRepoPath"] = connectWorkspaceRepoPath
	m["disconnectWorkspaceRepoPath"] = disconnectWorkspaceRepoPath
	m["startRunPath"] = startRunPath
}
