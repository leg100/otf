package e2e

import (
	"context"
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
		var gotGithubLocation string
		var gotOTFCallback string

		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate("https://localhost:8080"),
			chromedp.Text(".center", &gotLoginPrompt, chromedp.NodeVisible),
			chromedp.Click(".center > a", chromedp.NodeVisible),
		})
		require.NoError(t, err)

		err = chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Location(&gotGithubLocation),
		})
		require.NoError(t, err)

		err = chromedp.Run(ctx, chromedp.Tasks{
			chromedp.WaitVisible(`#login_field`, chromedp.ByID),
			chromedp.Focus(`#login_field`, chromedp.ByID),
			input.InsertText(username),
		})
		require.NoError(t, err)

		err = chromedp.Run(ctx, chromedp.Tasks{
			chromedp.WaitVisible(`#password`, chromedp.ByID),
			chromedp.Focus(`#password`, chromedp.ByID),
			input.InsertText(password),
		})
		require.NoError(t, err)

		err = chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Submit(`#password`, chromedp.ByID),
		})
		require.NoError(t, err)

		err = chromedp.Run(ctx, chromedp.Tasks{
			chromedp.WaitVisible(`.logo`),
			chromedp.Location(&gotOTFCallback),
		})
		require.NoError(t, err)

		assert.Equal(t, "Login with Github", strings.TrimSpace(gotLoginPrompt))
		assert.Regexp(t, `^https://github.com/login`, gotGithubLocation)
		assert.Regexp(t, `https://localhost:8080/organizations`, gotOTFCallback)
	})
}
