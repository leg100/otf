package integration

import (
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/daemon"
)

// TestSiteAdminUI demonstrates signing into the UI as a site admin
func TestSiteAdminUI(t *testing.T) {
	integrationTest(t)

	daemon, _, _ := setup(t, &config{Config: daemon.Config{
		SiteToken: "abc123",
	}})

	// nil ctx skips seeding browser with a session cookie
	browser.New(t, nil, chromedp.Tasks{
		// login as site admin
		_, err = page.Goto("https://" + daemon.System.Hostname() + "/login")
require.NoError(t, err)
		//screenshot(t, "no_authenticators_site_admin_login"),
		// use the link in the bottom right corner
		matchText(t, ".footer-site-login", "site admin", chromedp.ByQuery),
		err := page.Locator(".footer-site-login > a").Click()
require.NoError(t, err)
		//screenshot(t),
		// enter token
		chromedp.Focus("input#token", chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("abc123"),
		//screenshot(t, "site_admin_login_enter_token"),
		chromedp.Submit("input#token", chromedp.ByQuery),
		//screenshot(t, "site_admin_profile"),
		matchText(t, "#content > p", "You are logged in as site-admin", chromedp.ByQuery),
	})
}
