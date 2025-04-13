package integration

import (
	"testing"

	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/user"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestGithubLogin demonstrates logging into the UI via Github OAuth.
func TestGithubLogin(t *testing.T) {
	integrationTest(t)

	// Start daemon with a stub github server populated with a user.
	cfg := config{
		Config: daemon.Config{
			// specifying oauth credentials turns on the option to login via
			// github
			GithubClientID:     "stub-client-id",
			GithubClientSecret: "stub-client-secret",
		},
	}
	username := user.MustUsername("bobby")
	svc, _, _ := setup(t, &cfg, github.WithUsername(username))

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
