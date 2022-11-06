package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWeb is a random walkthrough of the Web UI
func TestWeb(t *testing.T) {
	addBuildsToPath(t)

	org := otf.NewTestOrganization(t)
	owners := otf.NewTeam("owners", org)
	devops := otf.NewTeam("devops", org)
	user := otf.NewTestUser(t, otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(owners, devops))
	hostname := startDaemon(t, user)
	url := "https://" + hostname

	// TODO: move tests out of subtests - we're not testing bits of functionality
	// independently but serially.

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
		createWebWorkspace(t, allocator, url, org.Name())
	})

	t.Run("assign workspace manager role to devops team", func(t *testing.T) {
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

	t.Run("add write workspace permission to owners team", func(t *testing.T) {
		workspace := createWebWorkspace(t, allocator, url, org.Name())

		// assign write permissions to team
		addWorkspacePermission(t, allocator, url, org.Name(), workspace, devops.Name(), "write")
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
