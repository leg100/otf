package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/daemon"
	"github.com/stretchr/testify/require"
)

// TestIntegration_OIDC demonstrates logging in using OIDC
func TestIntegration_OIDC(t *testing.T) {
	t.Parallel()

	// Start daemon with a stub github server populated with a user.
	cfg := config{
		Config: daemon.Config{
			OIDC: cloud.OIDCConfig{
				Name:                "google",
				IssuerURL:           authenticator.NewOIDCIssuer(t, "bobby", "stub-client-id", "google"),
				ClientID:            "stub-client-id",
				ClientSecret:        "stub-client-secret",
				SkipTLSVerification: true,
			},
		},
	}

	svc := setup(t, &cfg)

	browser := createBrowserCtx(t)
	err := chromedp.Run(browser, chromedp.Tasks{
		// go to login page
		chromedp.Navigate("https://" + svc.Hostname() + "/login"),
		screenshot(t, "oidc_login_button"),
		// login
		chromedp.Click("a.login-button-google", chromedp.NodeVisible),
		screenshot(t),
		// check login confirmation message
		matchText(t, ".content > p", "You are logged in as bobby"),
	})
	require.NoError(t, err)
}
