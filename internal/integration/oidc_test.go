package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/daemon"
)

// TestIntegration_OIDC demonstrates logging in using OIDC
func TestIntegration_OIDC(t *testing.T) {
	integrationTest(t)

	// Start daemon with a stub github server populated with a user.
	cfg := config{
		Config: daemon.Config{
			OIDC: authenticator.OIDCConfig{
				Name:                "google",
				IssuerURL:           authenticator.NewOIDCIssuer(t, "bobby", "stub-client-id", "google"),
				ClientID:            "stub-client-id",
				ClientSecret:        "stub-client-secret",
				SkipTLSVerification: true,
				UsernameClaim:       string(authenticator.DefaultUsernameClaim),
			},
		},
	}

	svc, _, _ := setup(t, &cfg)

	browser.New(t, nil, chromedp.Tasks{
		// go to login page
		_, err = page.Goto("https://" + svc.System.Hostname() + "/login")
require.NoError(t, err)
		//screenshot(t, "oidc_login_button"),
		// login
		err := page.Locator("a#login-button-google").Click()
require.NoError(t, err)
		//screenshot(t),
		// check login confirmation message
		matchText(t, "#content > p", "You are logged in as bobby", chromedp.ByQuery),
	})
}
