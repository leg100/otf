package integration

import (
	"testing"

	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/github"
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
	username := "bobby"
	svc, _, _ := setup(t, &cfg, github.WithUser(&username))

	page := browser.New(t, nil)

	// go to login page
	_, err := page.Goto("https://" + svc.System.Hostname() + "/login")
	require.NoError(t, err)
	//screenshot(t, "github_login_button"),

	// login
	err = page.Locator("a#login-button-github").Click()
	require.NoError(t, err)
	//screenshot(t),

	// check login confirmation message
	err = expect.Locator(page.Locator(`#content > p`)).ToHaveText(`You are logged in as bobby`)
	require.NoError(t, err)
}
