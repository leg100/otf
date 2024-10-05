package integration

import (
	"fmt"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/team"
	"github.com/stretchr/testify/require"
)

// TestWeb is a random walkthrough of the Web UI
func TestWeb(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)
	user := userFromContext(t, ctx)

	team, err := daemon.Teams.Create(ctx, org.Name, team.CreateTeamOptions{
		Name: internal.String("devops"),
	})
	require.NoError(t, err)
	err = daemon.Users.AddTeamMembership(ctx, team.ID, []string{user.Username})
	require.NoError(t, err)

	page := browser.New(t, ctx)
		// create workspace
		createWorkspace(t, daemon.System.Hostname(), org.Name, "my-workspace"),
		// assign workspace manager role to devops team
		chromedp.Tasks{
			// go to org
			_, err = page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
require.NoError(t, err)
			//screenshot(t),
			// list teams
			err := page.Locator("#teams > a").Click()
require.NoError(t, err)
			//screenshot(t),
			// select devops team
			err := page.Locator("#item-team-devops").Click()
require.NoError(t, err)
			//screenshot(t),
			// tick checkbox for workspace manager role
			err := page.Locator("#manage_workspaces").Click()
require.NoError(t, err)
			// submit form
			chromedp.Submit("#manage_workspaces", chromedp.NodeVisible, chromedp.ByQuery),
			//screenshot(t, "team_permissions_added_workspace_manager"),
			// confirm permissions updated
			matchText(t, "//div[@role='alert']", "team permissions updated"),
		},
		// add write permission on workspace to devops team
		addWorkspacePermission(t, daemon.System.Hostname(), org.Name, "my-workspace", team.ID, "write"),
		// list users
		chromedp.Tasks{
			// go to org
			_, err = page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
require.NoError(t, err)
			//screenshot(t),
			// list users
			err := page.Locator("#users > a").Click()
require.NoError(t, err)
			//screenshot(t),
			matchText(t, fmt.Sprintf("#item-user-%s #username", user.Username), user.Username, chromedp.ByQuery),
		},
		// list team members
		chromedp.Tasks{
			// go to org
			_, err = page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
require.NoError(t, err)
			//screenshot(t),
			// list teams
			err := page.Locator("#teams > a").Click()
require.NoError(t, err)
			// select owners team
			err := page.Locator("#item-team-owners").Click()
require.NoError(t, err)
			//screenshot(t),
			matchText(t, fmt.Sprintf("#item-user-%s #username", user.Username), user.Username, chromedp.ByQuery),
		},
	})
}
