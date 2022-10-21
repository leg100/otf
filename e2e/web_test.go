package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeb(t *testing.T) {
	addBuildsToPath(t)

	org := otf.NewTestOrganization(t)
	owners := otf.NewTeam("owners", org)
	devops := otf.NewTeam("devops", org)
	user := otf.NewTestUser(t, otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(owners, devops))
	hostname := startDaemon(t, user)
	url := "https://" + hostname

	t.Run("login", func(t *testing.T) {
		ctx, cancel := chromedp.NewContext(allocator)
		defer cancel()

		var gotLoginPrompt string
		var gotLocationOrganizations string

		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(url),
			ss.screenshot(t),
			chromedp.Text(".center", &gotLoginPrompt, chromedp.NodeVisible),
			chromedp.Click(".login-button-github", chromedp.NodeVisible),
			ss.screenshot(t),
			chromedp.Location(&gotLocationOrganizations),
		})
		require.NoError(t, err)

		assert.Equal(t, "Login with Github", strings.TrimSpace(gotLoginPrompt))
		assert.Equal(t, url+"/organizations", gotLocationOrganizations)
	})

	t.Run("new workspace", func(t *testing.T) {
		createWebWorkspace(t, allocator, url, org)
	})

	t.Run("assign workspace manager role to team", func(t *testing.T) {
		ctx, cancel := chromedp.NewContext(allocator)
		defer cancel()

		var gotFlashSuccess string
		orgSelector := fmt.Sprintf("#item-organization-%s a", org.Name())
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(url),
			// login
			chromedp.Click(".login-button-github", chromedp.NodeVisible, chromedp.ByQuery),
			// select org
			chromedp.Click(orgSelector, chromedp.NodeVisible, chromedp.ByQuery),
			// list teams
			chromedp.Click("#teams > a", chromedp.NodeVisible, chromedp.ByQuery),
			// select devops team
			chromedp.Click("#item-team-devops a", chromedp.NodeVisible, chromedp.ByQuery),
			ss.screenshot(t),
			// tick checkbox for workspace manager role
			chromedp.Click("#manage_workspaces", chromedp.NodeVisible, chromedp.ByQuery),
			// submit form
			chromedp.Submit("#manage_workspaces", chromedp.NodeVisible, chromedp.ByQuery),
			// capture flash message
			chromedp.Text(".flash-success", &gotFlashSuccess, chromedp.NodeVisible, chromedp.ByQuery),
		})
		require.NoError(t, err)

		assert.Equal(t, "team permissions updated", strings.TrimSpace(gotFlashSuccess))
	})

	t.Run("add workspace permission", func(t *testing.T) {
		workspace := createWebWorkspace(t, allocator, url, org)

		// assign write permissions to team
		addWorkspacePermission(t, allocator, url, org.Name(), workspace, owners.Name(), "write")
	})

	t.Run("list users", func(t *testing.T) {
		ctx, cancel := chromedp.NewContext(allocator)
		defer cancel()

		var gotUser string
		orgSelector := fmt.Sprintf("#item-organization-%s a", org.Name())
		userSelector := fmt.Sprintf("#item-user-%s .status", user.Username())
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(url),
			// login
			chromedp.Click(".login-button-github", chromedp.NodeVisible),
			// select org
			chromedp.Click(orgSelector, chromedp.NodeVisible),
			ss.screenshot(t),
			// list users
			chromedp.Click("#users > a", chromedp.NodeVisible),
			ss.screenshot(t),
			chromedp.Text(userSelector, &gotUser, chromedp.NodeVisible),
		})
		require.NoError(t, err)

		assert.Equal(t, user.Username(), strings.TrimSpace(gotUser))
	})

	t.Run("list team members", func(t *testing.T) {
		ctx, cancel := chromedp.NewContext(allocator)
		defer cancel()

		var gotUser string
		orgSelector := fmt.Sprintf("#item-organization-%s a", org.Name())
		userSelector := fmt.Sprintf("#item-user-%s .status", user.Username())
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(url),
			// login
			chromedp.Click(".login-button-github", chromedp.NodeVisible),
			// select org
			chromedp.Click(orgSelector, chromedp.NodeVisible),
			ss.screenshot(t),
			// list teams
			chromedp.Click("#teams > a", chromedp.NodeVisible),
			ss.screenshot(t),
			// select owners team
			chromedp.Click("#item-team-owners a", chromedp.NodeVisible),
			ss.screenshot(t),
			chromedp.Text(userSelector, &gotUser, chromedp.NodeVisible),
		})
		require.NoError(t, err)

		assert.Equal(t, user.Username(), strings.TrimSpace(gotUser))
	})
}

func createWebWorkspace(t *testing.T, ctx context.Context, url string, org *otf.Organization) string {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var gotFlashSuccess string
	workspaceName := "workspace-" + otf.GenerateRandomString(4)
	orgSelector := fmt.Sprintf("#item-organization-%s a", org.Name())

	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Click(".login-button-github", chromedp.NodeVisible),
		chromedp.Click(orgSelector, chromedp.NodeVisible),
		ss.screenshot(t),
		chromedp.Click("#menu-item-workspaces > a", chromedp.ByQuery),
		chromedp.Click("#new-workspace-button", chromedp.NodeVisible, chromedp.ByQuery),
		ss.screenshot(t),
		chromedp.Focus("input#name", chromedp.NodeVisible),
		input.InsertText(workspaceName),
		chromedp.Click("#create-workspace-button"),
		ss.screenshot(t),
		chromedp.Text(".flash-success", &gotFlashSuccess, chromedp.NodeVisible),
	})
	require.NoError(t, err)

	assert.Equal(t, "created workspace: "+workspaceName, strings.TrimSpace(gotFlashSuccess))

	return workspaceName
}

// addWorkspacePermission adds a workspace permission via the web app, assigning
// a role to a team on a workspace in an org.
func addWorkspacePermission(t *testing.T, allocater context.Context, url, org, workspace, team, role string) {
	ctx, cancel := chromedp.NewContext(allocater)
	defer cancel()

	var gotOwnersTeam string
	var gotOwnersRole string
	var gotFlashSuccess string

	orgSelector := fmt.Sprintf("#item-organization-%s a", org)
	workspaceSelector := fmt.Sprintf("#item-workspace-%s a", workspace)
	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		// login
		chromedp.Click(".login-button-github", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		// select org
		chromedp.Click(orgSelector, chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		// list workspaces
		chromedp.Click("#menu-item-workspaces > a", chromedp.NodeVisible, chromedp.ByQuery),
		chromedp.WaitReady(`body`),
		// select workspace
		chromedp.Click(workspaceSelector, chromedp.NodeVisible),
		ss.screenshot(t),
		// confirm builtin admin permission for owners team
		chromedp.Text("#permissions-owners td:first-child", &gotOwnersTeam, chromedp.NodeVisible),
		chromedp.Text("#permissions-owners td:last-child", &gotOwnersRole, chromedp.NodeVisible),
		// add write permission for the test team
		chromedp.SetValue(`//select[@id="permissions-add-select-role"]`, "write", chromedp.BySearch),
		chromedp.SetValue(`//select[@id="permissions-add-select-team"]`, team, chromedp.BySearch),
		chromedp.Click("#permissions-add-button", chromedp.NodeVisible),
		ss.screenshot(t),
		chromedp.Text(".flash-success", &gotFlashSuccess, chromedp.NodeVisible),
	})
	require.NoError(t, err)

	assert.Equal(t, "owners", gotOwnersTeam)
	assert.Equal(t, "admin", gotOwnersRole)
	assert.Equal(t, "updated workspace permissions", gotFlashSuccess)
}
