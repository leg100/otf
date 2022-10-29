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

		var footerLoginText, loginConfirmation string

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
		})
		require.NoError(t, err)

		assert.Equal(t, "site admin", footerLoginText)
		assert.Equal(t, "You are logged in as site-admin", strings.TrimSpace(loginConfirmation))
	})
}
