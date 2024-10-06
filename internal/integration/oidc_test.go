package integration

import (
	"testing"

	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/daemon"
	"github.com/stretchr/testify/require"
)

// TestIntegration_OIDC demonstrates logging in using OIDC
func TestIntegration_OIDC(t *testing.T) {
	integrationTest(t)

	// Start daemon configured to use a google OIDC test stub.
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

	page := browser.New(t, nil)

	// go to login page
	_, err := page.Goto("https://" + svc.System.Hostname() + "/login")
	require.NoError(t, err)
	//screenshot(t, "oidc_login_button"),

	// login
	err = page.Locator("a#login-button-google").Click()
	require.NoError(t, err)
	page.Pause()

	// check login confirmation message
	err = expect.Locator(page.Locator("#content > p")).ToHaveText("You are logged in as bobby")
	require.NoError(t, err)
}
