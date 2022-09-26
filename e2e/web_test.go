package e2e

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeb(t *testing.T) {
	username := lookupEnv(t, "OTF_E2E_GITHUB_USERNAME")
	password := lookupEnv(t, "OTF_E2E_GITHUB_PASSWORD")
	headless, ok := os.LookupEnv("OTF_E2E_HEADLESS")
	if !ok {
		headless = "false"
	}

	// create context
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", headless),
			chromedp.Flag("hide-scrollbars", true),
			chromedp.Flag("mute-audio", true),
			chromedp.Flag("ignore-certificate-errors", true),
			chromedp.Flag("disable-gpu", true),
		)...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	t.Run("login", func(t *testing.T) {
		var gotLoginPrompt string
		var gotGithubLoginLocation string
		var gotOTFOrganizationsLocation string

		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate("https://localhost:8080"),
			screenshot("otf_login"),

			chromedp.Text(".center", &gotLoginPrompt, chromedp.NodeVisible),
			chromedp.Click(".center > a", chromedp.NodeVisible),
			screenshot("github_login"),

			chromedp.Location(&gotGithubLoginLocation),
			chromedp.WaitVisible(`#login_field`, chromedp.ByID),
			chromedp.Focus(`#login_field`, chromedp.ByID),
			input.InsertText(username),
			chromedp.WaitVisible(`#password`, chromedp.ByID),
			chromedp.Focus(`#password`, chromedp.ByID),
			input.InsertText(password),
			screenshot("github_login_form_completed"),

			chromedp.Submit(`#password`, chromedp.ByID),
			screenshot("otf_login_successful"),

			chromedp.Location(&gotOTFOrganizationsLocation),
		})
		require.NoError(t, err)

		assert.Equal(t, "Login with Github", strings.TrimSpace(gotLoginPrompt))
		assert.Regexp(t, `^https://github.com/login`, gotGithubLoginLocation)
		assert.Regexp(t, `^https://localhost:8080/organizations`, gotOTFOrganizationsLocation)
	})
}

var screenshotCounter = 0

func screenshot(name string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		screenshotCounter++

		var image []byte
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.WaitReady(`body`),
			chromedp.CaptureScreenshot(&image),
		})
		if err != nil {
			return err
		}
		err = os.MkdirAll("screenshots", 0o755)
		if err != nil {
			return err
		}
		err = os.WriteFile(fmt.Sprintf("screenshots/%02d_%s.png", screenshotCounter, name), image, 0o644)
		if err != nil {
			return err
		}
		return nil
	}
}
