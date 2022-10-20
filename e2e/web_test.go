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
	allocater := newBrowserAllocater(t)

	org := otf.NewTestOrganization(t)
	team := otf.NewTestTeam(t, org)
	user := otf.NewTestUser(t, otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(team))
	hostname := startDaemon(t, user)
	url := "https://" + hostname

	t.Run("login", func(t *testing.T) {
		s := screenshotter(0)

		ctx, cancel := chromedp.NewContext(allocater)
		defer cancel()

		var gotLoginPrompt string
		var gotLocationOrganizations string

		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(url),
			s.screenshot(t),
			chromedp.Text(".center", &gotLoginPrompt, chromedp.NodeVisible),
			chromedp.Click(".login-button-github", chromedp.NodeVisible),
			s.screenshot(t),
			chromedp.Location(&gotLocationOrganizations),
		})
		require.NoError(t, err)

		assert.Equal(t, "Login with Github", strings.TrimSpace(gotLoginPrompt))
		assert.Equal(t, url+"/organizations", gotLocationOrganizations)
	})

	t.Run("new workspace", func(t *testing.T) {
		s := screenshotter(0)

		createWebWorkspace(t, allocater, s, url, org)
	})

	t.Run("add workspace permission", func(t *testing.T) {
		s := screenshotter(0)

		workspace := createWebWorkspace(t, allocater, s, url, org)

		// assign write permissions to team
		addWorkspacePermission(t, allocater, url, org.Name(), workspace, team.Name(), "write")
	})

	t.Run("list users", func(t *testing.T) {
		s := screenshotter(0)

		ctx, cancel := chromedp.NewContext(allocater)
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
			s.screenshot(t),
			// list users
			chromedp.Click("#users > a", chromedp.NodeVisible),
			s.screenshot(t),
			chromedp.Text(userSelector, &gotUser, chromedp.NodeVisible),
		})
		require.NoError(t, err)

		assert.Equal(t, user.Username(), strings.TrimSpace(gotUser))
	})

	t.Run("list team members", func(t *testing.T) {
		s := screenshotter(0)

		ctx, cancel := chromedp.NewContext(allocater)
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
			s.screenshot(t),
			// list teams
			chromedp.Click("#teams > a", chromedp.NodeVisible),
			s.screenshot(t),
			// select owners team
			chromedp.Click("#item-team-owners a", chromedp.NodeVisible),
			s.screenshot(t),
			chromedp.Text(userSelector, &gotUser, chromedp.NodeVisible),
		})
		require.NoError(t, err)

		assert.Equal(t, user.Username(), strings.TrimSpace(gotUser))
	})
}

func createWebWorkspace(t *testing.T, ctx context.Context, s screenshotter, url string, org *otf.Organization) string {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var gotFlashSuccess string
	workspaceName := "workspace-" + otf.GenerateRandomString(4)
	orgSelector := fmt.Sprintf("#item-organization-%s a", org.Name())

	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Click(".login-button-github", chromedp.NodeVisible),
		chromedp.Click(orgSelector, chromedp.NodeVisible),
		s.screenshot(t),
		chromedp.Click("#workspaces > a", chromedp.NodeVisible),
		chromedp.Click("#new-workspace-button", chromedp.NodeVisible),
		s.screenshot(t),
		chromedp.Focus("input#name", chromedp.NodeVisible),
		input.InsertText(workspaceName),
		chromedp.Click("#create-workspace-button"),
		s.screenshot(t),
		chromedp.Text(".flash-success", &gotFlashSuccess, chromedp.NodeVisible),
	})
	require.NoError(t, err)

	assert.Equal(t, "created workspace: "+workspaceName, strings.TrimSpace(gotFlashSuccess))

	return workspaceName
}

// addWorkspacePermission adds a workspace permission via the web app, assigning
// a role to a team on a workspace in an org.
func addWorkspacePermission(t *testing.T, allocater context.Context, url, org, workspace, team, role string) {
	s := screenshotter(0)

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
		chromedp.Click("#workspaces > a", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		// select workspace
		chromedp.Click(workspaceSelector, chromedp.NodeVisible),
		s.screenshot(t),
		// confirm builtin admin permission for owners team
		chromedp.Text("#permissions-owners td:first-child", &gotOwnersTeam, chromedp.NodeVisible),
		chromedp.Text("#permissions-owners td:last-child", &gotOwnersRole, chromedp.NodeVisible),
		// add write permission for the test team
		chromedp.SetValue(`//select[@id="permissions-add-select-role"]`, "write", chromedp.BySearch),
		chromedp.SetValue(`//select[@id="permissions-add-select-team"]`, team, chromedp.BySearch),
		chromedp.Click("#permissions-add-button", chromedp.NodeVisible),
		s.screenshot(t),
		chromedp.Text(".flash-success", &gotFlashSuccess, chromedp.NodeVisible),
	})
	require.NoError(t, err)

	assert.Equal(t, "owners", gotOwnersTeam)
	assert.Equal(t, "admin", gotOwnersRole)
	assert.Equal(t, "updated workspace permissions", gotFlashSuccess)
}
