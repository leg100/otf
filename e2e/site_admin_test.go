package e2e

import (
	"strings"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSiteAdmin demonstrates signing into the web app as a site admin.
func TestSiteAdmin(t *testing.T) {
	addBuildsToPath(t)

	// Create test user merely because startDaemon expects one, but this test
	// doesn't use it.
	user := otf.NewTestUser(t)
	hostname := startDaemon(t, user, "--site-token", "abc123")

	t.Run("login", func(t *testing.T) {
		ctx, cancel := chromedp.NewContext(allocator)
		defer cancel()

		var footerLoginText, loginConfirmation, orgCreated, orgLocation string

		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate("https://" + hostname + "/login"),
			ss.screenshot(t),
			chromedp.Text(".footer-site-login", &footerLoginText, chromedp.NodeVisible),
			chromedp.Click(".footer-site-login > a", chromedp.NodeVisible),
			ss.screenshot(t),
			chromedp.Focus("input#token", chromedp.NodeVisible),
			input.InsertText("abc123"),
			ss.screenshot(t),
			chromedp.Submit("input#token"),
			ss.screenshot(t),
			chromedp.Text(".content > p", &loginConfirmation, chromedp.NodeVisible),
			// now go to the list of organizations
			chromedp.Navigate("https://" + hostname + "/organizations"),
			// add an org
			chromedp.Click("#new-organization-button", chromedp.NodeVisible),
			ss.screenshot(t),
			chromedp.Focus("input#name", chromedp.NodeVisible),
			input.InsertText("my-new-org"),
			ss.screenshot(t),
			chromedp.Submit("input#name"),
			ss.screenshot(t),
			chromedp.Location(&orgLocation),
			chromedp.Text(".flash-success", &orgCreated, chromedp.NodeVisible),
		})
		require.NoError(t, err)

		assert.Equal(t, "site admin", footerLoginText)
		assert.Equal(t, "You are logged in as site-admin", strings.TrimSpace(loginConfirmation))
		assert.Equal(t, "https://"+hostname+"/organizations/my-new-org", orgLocation)
		assert.Equal(t, "created organization: my-new-org", strings.TrimSpace(orgCreated))
	})
}
