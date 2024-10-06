package integration

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIntegration_TeamUI demonstrates managing teams and team members via the
// UI.
func TestIntegration_TeamUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)
	_, err := daemon.Users.Create(ctx, "bob")
	require.NoError(t, err)
	_, err = daemon.Users.Create(ctx, "alice")
	require.NoError(t, err)

	page := browser.New(t, ctx)

	// go to org
	_, err = page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
	require.NoError(t, err)
	//screenshot(t),
	// go to teams listing
	err = page.Locator(`//a[text()='teams']`).Click()
	require.NoError(t, err)
	//screenshot(t),
	// go to owners team page
	err = page.Locator(`//div[@id='item-team-owners']`).Click()
	require.NoError(t, err)
	//screenshot(t, "owners_team_page"),

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
	//screenshot(t),

	// submit
	err = page.Locator(`//input[@x-ref='input-search']`).Press(`Enter`)
	require.NoError(t, err)
	//screenshot(t),

	// confirm bob added
	err = expect.Locator(page.Locator("//div[@role='alert']")).ToHaveText("added team member: bob")
	require.NoError(t, err)

	// remove bob from team
	err = page.Locator(`//div[@id='item-user-bob']//button[@id='remove-member-button']`).Click()
	require.NoError(t, err)
	//screenshot(t),

	// confirm bob removed
	err = expect.Locator(page.Locator("//div[@role='alert']")).ToHaveText("removed team member: bob")
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
	//screenshot(t),

	// confirm sarah added
	err = expect.Locator(page.Locator("//div[@role='alert']")).ToHaveText("added team member: sarah")
	require.NoError(t, err)
}
