package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestWeb(t *testing.T) {
	headless, ok := os.LookupEnv("OTF_E2E_HEADLESS")
	if !ok {
		headless = "false"
	}

	githubHostname := githubStub(t)
	t.Setenv("OTF_GITHUB_HOSTNAME", githubHostname)

	startDaemon(t, 8002)

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
		var gotOTFOrganizationsLocation string

		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate("https://localhost:8002"),

			screenshot("otf_login"),

			chromedp.Text(".center", &gotLoginPrompt, chromedp.NodeVisible),
			chromedp.Click(".center > a", chromedp.NodeVisible),

			screenshot("otf_login_successful"),

			chromedp.Location(&gotOTFOrganizationsLocation),
		})
		require.NoError(t, err)

		assert.Equal(t, "Login with Github", strings.TrimSpace(gotLoginPrompt))
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

func githubStub(t *testing.T) string {
	authCode := otf.GenerateRandomString(10)

	http.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		q := url.Values{}
		q.Add("state", r.URL.Query().Get("state"))
		q.Add("code", authCode)

		// TODO: check if can use referrer header?
		callback := url.URL{
			Scheme:   "https",
			Host:     "localhost:8002",
			RawQuery: q.Encode(),
		}

		http.Redirect(w, r, callback.String(), 302)
	})
	http.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		token := oauth2.Token{AccessToken: "stub-token"}
		out, err := json.Marshal(&token)
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	srv := httptest.NewTLSServer(nil)
	defer srv.Close()
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	return u.Host
}
