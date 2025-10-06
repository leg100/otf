package integration

import (
	"regexp"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_TeamUI demonstrates managing teams and team members via the
// UI.
func TestIntegration_TeamUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)
	_, err := daemon.Users.Create(ctx, "bob")
	require.NoError(t, err)
	_, err = daemon.Users.Create(ctx, "alice")
	require.NoError(t, err)

	browser.New(t, ctx, func(page playwright.Page) {
		// go to org
		_, err = page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
		require.NoError(t, err)
		// go to teams listing
		err = page.Locator(`#menu-item-teams > a`).Click()
		require.NoError(t, err)
		// go to owners team page
		err = page.Locator(`//tr[@id='item-team-owners']/td[1]/a`).Click()
		require.NoError(t, err)
		screenshot(t, page, "owners_team_page")

		// set focus to search box
		err = page.Locator(`//input[@x-ref='input-search']`).Fill("")
		require.NoError(t, err)
		// input.InsertText(""),

		// should trigger dropdown box showing both alice and bob
		err = expect.Locator(page.Locator(`//div[@x-ref='searchdrop']//button[text()='bob']`)).ToBeVisible()
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//div[@x-ref='searchdrop']//button[text()='alice']`)).ToBeVisible()
		require.NoError(t, err)

		// select bob as new team member
		err = page.Locator(`//input[@x-ref='input-search']`).Fill("bob")
		require.NoError(t, err)

		// submit
		err = page.Locator(`//input[@x-ref='input-search']`).Press(`Enter`)
		require.NoError(t, err)

		// confirm bob added
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("added team member: bob")
		require.NoError(t, err)

		// remove bob from team
		err = page.Locator(`//*[@id='item-user-bob']//button[@id='remove-member-button']`).Click()
		require.NoError(t, err)

		// confirm bob removed
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("removed team member: bob")
		require.NoError(t, err)

		// now demonstrate specifying a username that doesn't belong to an
		// existing user. The dropdown box should prompt to create the user
		// and add them to the team.
		err = page.Locator(`//input[@x-ref='input-search']`).Fill("sarah")
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//div[@x-ref='searchdrop']//button`)).ToHaveText(regexp.MustCompile(`Create:.*sarah`))
		require.NoError(t, err)

		// submit
		err = page.Locator(`//input[@x-ref='input-search']`).Press(`Enter`)
		require.NoError(t, err)

		// confirm sarah added
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("added team member: sarah")
		require.NoError(t, err)
	})
}

func TestIntegration_TeamUI_Permissions(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)
	team := daemon.createTeam(t, ctx, org)

	browser.New(t, ctx, func(page playwright.Page) {
		// go to team's page
		teamURL := "https://" + daemon.System.Hostname() + "/app/teams/" + team.ID.String()
		_, err := page.Goto(teamURL)
		require.NoError(t, err)

		// new team should have no permissions
		err = expect.Locator(page.Locator(`//*[@id="manage_workspaces"]`)).Not().ToBeChecked()
		require.NoError(t, err)
		err = expect.Locator(page.Locator(`//*[@id="manage_vcs"]`)).Not().ToBeChecked()
		require.NoError(t, err)
		err = expect.Locator(page.Locator(`//*[@id="manage_modules"]`)).Not().ToBeChecked()
		require.NoError(t, err)

		// assign manage workspaces permission
		err = page.Locator(`//*[@id="manage_workspaces"]`).Check()
		require.NoError(t, err)

		// save changes
		err = page.Locator(`//*[@id="content"]/form[1]/div[4]/button`).Click()
		require.NoError(t, err)

		// expect flash message
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("team permissions updated")
		require.NoError(t, err)

		// manage workspaces permission checkbox should be checked
		err = expect.Locator(page.Locator(`//*[@id="manage_workspaces"]`)).ToBeChecked()
		require.NoError(t, err)

		// unassign manage workspaces permission
		err = page.Locator(`//*[@id="manage_workspaces"]`).Uncheck()
		require.NoError(t, err)

		// save changes
		err = page.Locator(`//*[@id="content"]/form[1]/div[4]/button`).Click()
		require.NoError(t, err)

		// expect flash message
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("team permissions updated")
		require.NoError(t, err)

		// manage workspaces permission checkbox should be unchecked
		err = expect.Locator(page.Locator(`//*[@id="manage_workspaces"]`)).Not().ToBeChecked()
		require.NoError(t, err)
	})
}
