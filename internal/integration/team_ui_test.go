package integration

import (
	"fmt"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

// TestIntegration_TeamUI demonstrates managing teams and team members via the
// UI.
func TestIntegration_TeamUI(t *testing.T) {
	t.Parallel()

	svc := setup(t, nil)
	owner, ownerCtx := svc.createUserCtx(t, ctx)
	org := svc.createOrganization(t, ownerCtx)
	newbie := svc.createUser(t, ctx)

	browser := createBrowserCtx(t)
	okDialog(t, browser)
	err := chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ownerCtx, svc.Hostname(), owner.Username, svc.Secret),
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(organizationURL(svc.Hostname(), org.Name)),
			screenshot(t),
			// go to teams listing
			chromedp.Click(`//a[text()='teams']`, chromedp.NodeVisible),
			screenshot(t),
			// go to owners team page
			chromedp.Click(`//div[@class='content-list']//a[text()='owners']`, chromedp.NodeVisible),
			screenshot(t),
			// select newbie as new team member
			chromedp.SetValue(`//select[@id="select-add-member"]`, newbie.Username, chromedp.BySearch),
			screenshot(t),
			// submit
			chromedp.Click(`//button[text()='Add member']`, chromedp.NodeVisible),
			screenshot(t),
			// confirm newbie added
			matchText(t, ".flash-success", "added team member: "+newbie.Username),
			// remove newbie from team
			chromedp.Click(fmt.Sprintf(`//div[@id='item-user-%s']//button[@id='remove-member-button']`, newbie.Username), chromedp.NodeVisible),
			screenshot(t),
			// confirm newbie removed
			matchText(t, ".flash-success", "removed team member: "+newbie.Username),
		},
	})
	require.NoError(t, err)
}
