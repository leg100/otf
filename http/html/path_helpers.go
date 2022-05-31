package html

import "fmt"

// TODO: autogenerate path helpers from router

// Profile path helpers

type pathHelpers struct{}

func (h pathHelpers) getProfilePath() string {
	return "/profile"
}

func (h pathHelpers) listSessionPath() string {
	return "/profile/sessions"
}

func (h pathHelpers) revokeSessionPath() string {
	return "/profile/revoke"
}

func (h pathHelpers) listTokenPath() string {
	return "/profile/tokens"
}

func (h pathHelpers) deleteTokenPath() string {
	return "/profile/tokens/delete"
}

func (h pathHelpers) newTokenPath() string {
	return "/profile/tokens/new"
}

func (h pathHelpers) createTokenPath() string {
	return "/profile/tokens/create"
}

// Workspace path helpers

func (h pathHelpers) listOrganizationPath() string {
	return "/organizations"
}

func (h pathHelpers) newOrganizationPath() string {
	return "/organizations/new"
}

func (h pathHelpers) createOrganizationPath() string {
	return "/organizations/create"
}

func (h pathHelpers) getOrganizationPath(organization_name string) string {
	return fmt.Sprintf("/organizations/%s", organization_name)
}

func (h pathHelpers) editOrganizationPath(organization_name string) string {
	return fmt.Sprintf("/organizations/%s/edit", organization_name)
}

func updateOrganizationPath(organization_name string) string {
	return fmt.Sprintf("/organizations/%s/update", organization_name)
}

func deleteOrganizationPath(organization_name string) string {
	return fmt.Sprintf("/organizations/%s/delete", organization_name)
}

// Workspace path helpers

func listWorkspacePath(organization_name string) string {
	return fmt.Sprintf("/organizations/%s/workspaces", organization_name)
}

func newWorkspacePath(organization_name string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/new", organization_name)
}

func createWorkspacePath(organization_name string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/create", organization_name)
}

func getWorkspacePath(organization_name, workspace_name string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s", organization_name, workspace_name)
}

func editWorkspacePath(organization_name, workspace_name string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/edit", organization_name, workspace_name)
}

func updateWorkspacePath(organization_name, workspace_name string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/update", organization_name, workspace_name)
}

func deleteWorkspacePath(organization_name, workspace_name string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/delete", organization_name, workspace_name)
}

func lockWorkspacePath(organization_name, workspace_name string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/lock", organization_name, workspace_name)
}

func unlockWorkspacePath(organization_name, workspace_name string) string {
	return fmt.Sprintf("/organizations/%s/workspaces/%s/unlock", organization_name, workspace_name)
}
