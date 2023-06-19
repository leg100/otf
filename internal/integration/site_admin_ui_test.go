package integration

import (
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSiteAdminUI demonstrates signing into the web app as a site admin, using
// their super powers to create and delete an organization.
func TestSiteAdminUI(t *testing.T) {
	t.Parallel()

	daemon := setup(t, &config{Config: daemon.Config{
		SiteToken: "abc123",
	}})

	var orgLocation string

	browser := createTabCtx(t)
	// Click OK on any browser javascript dialog boxes that pop up
	okDialog(t, browser)
	err := chromedp.Run(browser, chromedp.Tasks{
		// login as site admin
		chromedp.Navigate("https://" + daemon.Hostname() + "/login"),
		screenshot(t, "no_authenticators_site_admin_login"),
		// use the link in the bottom right corner
		matchText(t, ".footer-site-login", "site admin"),
		chromedp.Click(".footer-site-login > a", chromedp.NodeVisible),
		screenshot(t),
		// enter token
		chromedp.Focus("input#token", chromedp.NodeVisible),
		input.InsertText("abc123"),
		screenshot(t, "site_admin_login_enter_token"),
		chromedp.Submit("input#token"),
		screenshot(t, "site_admin_profile"),
		matchText(t, ".content > p", "You are logged in as site-admin"),
		// now go to the list of organizations
		chromedp.Navigate("https://" + daemon.Hostname() + "/app/organizations"),
		// add an org
		chromedp.Click("#new-organization-button", chromedp.NodeVisible),
		screenshot(t),
		chromedp.Focus("input#name", chromedp.NodeVisible),
		input.InsertText("my-new-org"),
		screenshot(t, "new_org_enter_name"),
		chromedp.Submit("input#name"),
		screenshot(t, "new_org_created"),
		chromedp.Location(&orgLocation),
		matchText(t, ".flash-success", "created organization: my-new-org"),
		// go to organization settings
		chromedp.Click("#settings > a", chromedp.NodeVisible),
		screenshot(t),
		// change organization name
		chromedp.Focus("input#name", chromedp.NodeVisible),
		chromedp.Clear("input#name"),
		input.InsertText("newly-named-org"),
		screenshot(t),
		chromedp.Click(`//button[text()='Update organization name']`, chromedp.NodeVisible),
		screenshot(t),
		matchText(t, ".flash-success", "updated organization"),
		// delete the organization
		chromedp.Click(`//button[@id='delete-organization-button']`, chromedp.NodeVisible),
		screenshot(t),
		matchText(t, ".flash-success", "deleted organization: newly-named-org"),
	})
	require.NoError(t, err)

	assert.Equal(t, organizationURL(daemon.Hostname(), "my-new-org"), orgLocation)
}
