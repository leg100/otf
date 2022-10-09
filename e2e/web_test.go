package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestWeb(t *testing.T) {
	headless := true
	if v, ok := os.LookupEnv("OTF_E2E_HEADLESS"); ok {
		var err error
		headless, err = strconv.ParseBool(v)
		require.NoError(t, err)
	}

	githubHostname := githubStub(t)
	t.Setenv("OTF_GITHUB_HOSTNAME", githubHostname)

	url := startDaemon(t)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", headless),
			chromedp.Flag("hide-scrollbars", true),
			chromedp.Flag("mute-audio", true),
			chromedp.Flag("ignore-certificate-errors", true),
			chromedp.Flag("disable-gpu", true),
		)...)
	defer cancel()

	t.Run("login", func(t *testing.T) {
		ctx, cancel := chromedp.NewContext(allocCtx)
		defer cancel()

		var gotLoginPrompt string
		var gotLocationOrganizations string

		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(url),
			screenshot("otf_login"),
			chromedp.Text(".center", &gotLoginPrompt, chromedp.NodeVisible),
			chromedp.Click(".login-button-github", chromedp.NodeVisible),
			screenshot("otf_login_successful"),
			chromedp.Location(&gotLocationOrganizations),
		})
		require.NoError(t, err)

		assert.Equal(t, "Login with Github", strings.TrimSpace(gotLoginPrompt))
		assert.Equal(t, url+"/organizations", gotLocationOrganizations)
	})

	t.Run("new workspace", func(t *testing.T) {
		ctx, cancel := chromedp.NewContext(allocCtx)
		defer cancel()

		var gotFlashSuccess string
		workspaceName := "workspace-" + otf.GenerateRandomString(4)

		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(url),
			chromedp.Click(".login-button-github", chromedp.NodeVisible),
			chromedp.Click(".content-list a", chromedp.NodeVisible),
			chromedp.Click("#workspaces > a", chromedp.NodeVisible),
			chromedp.Click("#new-workspace-button", chromedp.NodeVisible),
			screenshot("otf_new_workspace_form"),
			chromedp.Focus("input#name", chromedp.NodeVisible),
			input.InsertText(workspaceName),
			chromedp.Click("#create-workspace-button"),
			screenshot("otf_created_workspace"),
			chromedp.Text(".flash-success", &gotFlashSuccess, chromedp.NodeVisible),
		})
		require.NoError(t, err)

		assert.Equal(t, "created workspace: "+workspaceName, strings.TrimSpace(gotFlashSuccess))
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

	http.HandleFunc("/login/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		q := url.Values{}
		q.Add("state", r.URL.Query().Get("state"))
		q.Add("code", authCode)

		referrer, err := url.Parse(r.Referer())
		require.NoError(t, err)

		// TODO: check if can use referrer header?
		callback := url.URL{
			Scheme:   referrer.Scheme,
			Host:     referrer.Host,
			Path:     "/oauth/github/callback",
			RawQuery: q.Encode(),
		}

		http.Redirect(w, r, callback.String(), http.StatusFound)
	})
	http.HandleFunc("/login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(&oauth2.Token{AccessToken: "stub_token"})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	http.HandleFunc("/api/v3/user", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(&github.User{Login: otf.String("stub_user")})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	http.HandleFunc("/api/v3/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal([]*github.Organization{{Login: otf.String("stub_org")}})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	http.HandleFunc("/api/v3/user/teams", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal([]*github.Team{
			{
				Name: otf.String("stub_team"),
				Organization: &github.Organization{
					Login: otf.String("stub_org"),
				},
			},
		})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	srv := httptest.NewTLSServer(nil)
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	return u.Host
}
