package integration

import (
	"testing"

	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/user"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestGithubLogin demonstrates logging into the UI via Github OAuth.
func TestGithubLogin(t *testing.T) {
	integrationTest(t)

	// Start daemon with a stub github server populated with a user.
	svc, _, _ := setup(t,
		// specifying oauth credentials turns on the option to login via
		// github
		withGithubOAuthCredentials("stub-client-id", "stub-client-secret"),
		withGithubOption(github.WithUsername(user.MustUsername("bobby"))),
	)

	browser.New(t, nil, func(page playwright.Page) {
		// go to login page
		_, err := page.Goto("https://" + svc.System.Hostname() + "/login")
		require.NoError(t, err)
		screenshot(t, page, "github_login_button")

		// login
		err = page.Locator("a#login-button-github").Click()
		require.NoError(t, err)

		// check login confirmation message
		err = expect.Locator(page.Locator(`//main`)).ToContainText(`You are logged in as bobby`)
		require.NoError(t, err)
	})
}
