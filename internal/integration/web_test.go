package integration

import (
	"fmt"
	"testing"

	"github.com/leg100/otf/internal/team"
	userpkg "github.com/leg100/otf/internal/user"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestWeb is a random walkthrough of the Web UI
func TestWeb(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)
	user := userFromContext(t, ctx)

	team, err := daemon.Teams.Create(ctx, org.Name, team.CreateTeamOptions{
		Name: new("devops"),
	})
	require.NoError(t, err)
	err = daemon.Users.AddTeamMembership(ctx, team.ID, []userpkg.Username{user.Username})
	require.NoError(t, err)

	browser.New(t, ctx, func(page playwright.Page) {
		// create workspace
		createWorkspace(t, page, daemon.System.Hostname(), org.Name, "my-workspace")
		// assign workspace manager role to devops team
		// go to org
		_, err = page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
		require.NoError(t, err)
		// list teams
		err = page.Locator("#menu-item-teams > a").Click()
		require.NoError(t, err)
		// select devops team
		err = page.Locator(`//tr[@id='item-team-devops']/td[1]/a`).Click()
		require.NoError(t, err)
		// tick checkbox for workspace manager role
		err = page.Locator("#manage_workspaces").Click()
		require.NoError(t, err)
		// submit form
		err = page.GetByRole("button").GetByText("Save changes").Click()
		require.NoError(t, err)
		screenshot(t, page, "team_permissions_added_workspace_manager")
		// confirm permissions updated
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("team permissions updated")
		require.NoError(t, err)
		// add write permission on workspace to devops team
		addWorkspacePermission(t, page, daemon.System.Hostname(), org.Name, "my-workspace", team.ID, "write")
		// list users

		// go to org
		_, err = page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
		require.NoError(t, err)

		// list users
		err = page.Locator("#menu-item-users > a").Click()
		require.NoError(t, err)

		err = expect.Locator(page.Locator(fmt.Sprintf("#item-user-%s #username", user.Username))).ToHaveText(user.Username.String())
		require.NoError(t, err)

		// list team members

		// go to org
		_, err = page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
		require.NoError(t, err)

		// list teams
		err = page.Locator("#menu-item-teams > a").Click()
		require.NoError(t, err)

		// select owners team
		err = page.Locator(`//tr[@id='item-team-owners']/td[1]/a`).Click()
		require.NoError(t, err)

		err = expect.Locator(page.Locator(fmt.Sprintf("#item-user-%s #username", user.Username))).ToHaveText(user.Username.String())
		require.NoError(t, err)
	})
}
