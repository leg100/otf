package integration

import (
	"testing"

	"github.com/leg100/otf/internal/daemon"
	"github.com/stretchr/testify/require"
)

// TestSiteAdminUI demonstrates signing into the UI as a site admin
func TestSiteAdminUI(t *testing.T) {
	integrationTest(t)

	daemon, _, _ := setup(t, &config{Config: daemon.Config{
		SiteToken: "abc123",
	}})

	// nil ctx skips seeding browser with a session cookie
	page := browser.New(t, nil)

	// login as site admin
	_, err := page.Goto("https://" + daemon.System.Hostname() + "/login")
	require.NoError(t, err)
	screenshot(t, page, "no_authenticators_site_admin_login")

	// use the link in the bottom right corner
	err = expect.Locator(page.Locator(".footer-site-login")).ToHaveText("site admin")
	require.NoError(t, err)

	err = page.Locator(".footer-site-login > a").Click()
	require.NoError(t, err)

	// enter token
	err = page.Locator("input#token").Fill("abc123")
	require.NoError(t, err)
	screenshot(t, page, "site_admin_login_enter_token")

	// submit
	err = page.GetByRole("button").GetByText("Login").Click()
	require.NoError(t, err)
	screenshot(t, page, "site_admin_profile")

	err = expect.Locator(page.Locator("#content > p")).ToHaveText("You are logged in as site-admin")
	require.NoError(t, err)
}
