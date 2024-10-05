package integration

import (
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
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
		chromedp.Tasks{
			// go to org
			_, err = page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
require.NoError(t, err)
			//screenshot(t),
			// go to teams listing
			err := page.Locator(`//a[text()='teams']`).Click()
require.NoError(t, err)
			//screenshot(t),
			// go to owners team page
			err := page.Locator(`//div[@id='item-team-owners']`).Click()
require.NoError(t, err)
			//screenshot(t, "owners_team_page"),
			// set focus to search box
			chromedp.Focus(`//input[@x-ref='input-search']`, chromedp.NodeVisible),
			input.InsertText(""),
			// should trigger dropdown box showing both alice and bob
			chromedp.WaitVisible(`//div[@x-ref='searchdrop']//button[text()='bob']`),
			chromedp.WaitVisible(`//div[@x-ref='searchdrop']//button[text()='alice']`),
			// select bob as new team member
			input.InsertText("bob"),
			//screenshot(t),
			// submit
			chromedp.Submit(`//input[@x-ref='input-search']`),
			//screenshot(t),
			// confirm bob added
			matchText(t, "//div[@role='alert']", "added team member: bob"),
			// remove bob from team
			err := page.Locator(`//div[@id='item-user-bob']//button[@id='remove-member-button']`).Click()
require.NoError(t, err)
			//screenshot(t),
			// confirm bob removed
			matchText(t, "//div[@role='alert']", "removed team member: bob"),
			// now demonstrate specifying a username that doesn't belong to an
			// existing user. The dropdown box should prompt to create the user
			// and add them to the team.
			chromedp.Focus(`//input[@x-ref='input-search']`, chromedp.NodeVisible),
			input.InsertText("sarah"),
			matchRegex(t, `//div[@x-ref='searchdrop']//button`, `Create:.*sarah`),
			// submit
			chromedp.Submit(`//input[@x-ref='input-search']`),
			//screenshot(t),
			// confirm sarah added
			matchText(t, "//div[@role='alert']", "added team member: sarah"),
		},
	})
}
