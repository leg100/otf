package integration

import (
	"testing"

	"github.com/leg100/otf/internal/authenticator"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_OIDC demonstrates logging in using OIDC
func TestIntegration_OIDC(t *testing.T) {
	integrationTest(t)

	// Start daemon configured to use a google OIDC test stub.
	svc, _, _ := setup(t, withOIDConfig(
		authenticator.OIDCConfig{
			Name:                "google",
			IssuerURL:           authenticator.NewOIDCIssuer(t, "bobby", "stub-client-id", "google"),
			ClientID:            "stub-client-id",
			ClientSecret:        "stub-client-secret",
			SkipTLSVerification: true,
			UsernameClaim:       string(authenticator.DefaultUsernameClaim),
		},
	))

	browser.New(t, nil, func(page playwright.Page) {
		// go to login page
		_, err := page.Goto("https://" + svc.System.Hostname() + "/login")
		require.NoError(t, err)
		screenshot(t, page, "oidc_login_button")

		// login
		err = page.Locator("a#login-button-google").Click()
		require.NoError(t, err)
		page.Pause()

		// check login confirmation message
		err = expect.Locator(page.Locator("//main")).ToContainText("You are logged in as bobby")
		require.NoError(t, err)
	})
}
