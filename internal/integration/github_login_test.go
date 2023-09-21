package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/github"
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
	user := cloud.User{Name: "bobby"}
	svc, _, _ := setup(t, &cfg, github.WithUser(&user))

	browser.Run(t, nil, chromedp.Tasks{
		// go to login page
		chromedp.Navigate("https://" + svc.Hostname() + "/login"),
		screenshot(t, "github_login_button"),
		// login
		chromedp.Click("a#login-button-github", chromedp.ByQuery),
		screenshot(t),
		// check login confirmation message
		matchText(t, `#content > p`, `You are logged in as bobby`, chromedp.ByQuery),
	})
}
