package integration

import (
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/daemon"
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

	browser := createBrowserCtx(t)
	// Click OK on any browser javascript dialog boxes that pop up
	okDialog(t, browser)
	err := chromedp.Run(browser, chromedp.Tasks{
		// login as site admin
		chromedp.Navigate("https://" + daemon.Hostname() + "/login"),
		screenshot(t),
		// use the link in the bottom right corner
		matchText(t, ".footer-site-login", "site admin"),
		chromedp.Click(".footer-site-login > a", chromedp.NodeVisible),
		screenshot(t),
		// enter token
		chromedp.Focus("input#token", chromedp.NodeVisible),
		input.InsertText("abc123"),
		screenshot(t),
		chromedp.Submit("input#token"),
		screenshot(t),
		matchText(t, ".content > p", "You are logged in as site-admin"),
		// now go to the list of organizations
		chromedp.Navigate("https://" + daemon.Hostname() + "/app/organizations"),
		// add an org
		chromedp.Click("#new-organization-button", chromedp.NodeVisible),
		screenshot(t),
		chromedp.Focus("input#name", chromedp.NodeVisible),
		input.InsertText("my-new-org"),
		screenshot(t),
		chromedp.Submit("input#name"),
		screenshot(t),
		chromedp.Location(&orgLocation),
		matchText(t, ".flash-success", "created organization: my-new-org"),
		// return to the list of organizations
		chromedp.Navigate("https://" + daemon.Hostname() + "/app/organizations"),
		// delete the organization
		chromedp.Click(`//button[text()='delete']`, chromedp.NodeVisible),
		screenshot(t),
		matchText(t, ".flash-success", "deleted organization: my-new-org"),
	})
	require.NoError(t, err)

	assert.Equal(t, organizationPath(daemon.Hostname(), "my-new-org"), orgLocation)
}
