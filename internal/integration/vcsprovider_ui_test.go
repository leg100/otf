package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	gogithub "github.com/google/go-github/v65/github"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/require"
)

// TestIntegration_VCSProviderUI demonstrates management of personal token vcs providers via
// the UI.
func TestIntegration_VCSProviderTokenUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)

	// create a vcs provider with a github personal access token
	page := browser.New(t, ctx)
		// go to org
		_, err = page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
require.NoError(t, err)
		//screenshot(t, "organization_main_menu"),
		// go to vcs providers
		err := page.Locator("#vcs_providers > a").Click()
require.NoError(t, err)
		//screenshot(t, "vcs_providers_list"),
		// click 'New Github VCS Provider' button
		err := page.Locator(`//button[text()='New Github VCS Provider (Personal Token)']`).Click()
require.NoError(t, err)
		//screenshot(t, "new_github_vcs_provider_form"),
		// enter fake github token
		chromedp.Focus("textarea#token", chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("fake-github-personal-token"),
		// submit form to create provider
		chromedp.Submit("textarea#token", chromedp.ByQuery),
		matchText(t, "//div[@role='alert']", `created provider: github \(token\)`),
		//screenshot(t, "vcs_provider_created_github_pat_provider"),
		// edit provider
		err := page.Locator(`//a[@id='edit-vcs-provider-link']`).Click()
require.NoError(t, err) waitLoaded,
		// give it a name
		chromedp.Focus("input#name", chromedp.ByQuery, chromedp.NodeVisible),
		input.InsertText("my-token"),
		err := page.Locator(`//button[text()='Update']`).Click()
require.NoError(t, err)
		matchText(t, "//div[@role='alert']", "updated provider: my-token"),
		// change token
		err := page.Locator(`//a[@id='edit-vcs-provider-link']`).Click()
require.NoError(t, err) waitLoaded,
		chromedp.Focus("textarea#token", chromedp.ByQuery, chromedp.NodeVisible),
		input.InsertText("my-updated-fake-github-personal-token"),
		err := page.Locator(`//button[text()='Update']`).Click()
require.NoError(t, err)
		matchText(t, "//div[@role='alert']", "updated provider: my-token"),
		// clear name
		err := page.Locator(`//a[@id='edit-vcs-provider-link']`).Click()
require.NoError(t, err) waitLoaded,
		chromedp.Focus("input#name", chromedp.ByQuery, chromedp.NodeVisible),
		chromedp.Clear("input#name", chromedp.ByQuery),
		err := page.Locator(`//button[text()='Update']`).Click()
require.NoError(t, err)
		matchText(t, "//div[@role='alert']", `updated provider: github \(token\)`),
		// delete token
		err := page.Locator(`//a[@id='edit-vcs-provider-link']`).Click()
require.NoError(t, err) waitLoaded,
		err := page.Locator(`//button[@id='delete-vcs-provider-button']`).Click()
require.NoError(t, err)
		matchText(t, "//div[@role='alert']", `deleted provider: github \(token\)`),
	})
}

// TestIntegration_VCSProviderAppUI demonstrates management of github app vcs
// providers via the UI.
func TestIntegration_VCSProviderAppUI(t *testing.T) {
	integrationTest(t)

	// create github stub server and return its hostname.
	githubHostname := func(t *testing.T) string {
		install := &gogithub.Installation{
			ID:         internal.Int64(123),
			Account:    &gogithub.User{Login: internal.String("leg100")},
			TargetType: internal.String("User"),
		}
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v3/app/installations", func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal([]*gogithub.Installation{install})
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		mux.HandleFunc("/api/v3/app/installations/123", func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal(install)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		mux.HandleFunc("/api/v3/installation/repositories", func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal(&gogithub.ListRepositories{
				Repositories: []*gogithub.Repository{{FullName: internal.String("leg100/otf-workspaces")}},
			})
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		stub := httptest.NewTLSServer(mux)
		t.Cleanup(stub.Close)

		u, err := url.Parse(stub.URL)
		require.NoError(t, err)
		return u.Host
	}(t)

	daemon, org, _ := setup(t, &config{Config: daemon.Config{GithubHostname: githubHostname}})

	// creating a github app requires site-admin role
	ctx := internal.AddSubjectToContext(context.Background(), &user.SiteAdmin)

	// create app
	_, err := daemon.GithubApp.CreateApp(ctx, github.CreateAppOptions{
		AppID:      123,
		Slug:       "otf-123",
		PrivateKey: string(testutils.ReadFile(t, "./fixtures/key.pem")),
	})
	require.NoError(t, err)

	// create github app vcs provider via UI.
	page := browser.New(t, ctx)
		// go to org
		_, err = page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
require.NoError(t, err)
		// go to vcs providers
		err := page.Locator("#vcs_providers > a").Click()
require.NoError(t, err)
		//screenshot(t, "vcs_provider_list_including_github_app"),
		// click button for creating a new vcs provider with a github app
		err := page.Locator(`//button[text()='New Github VCS Provider (App)']`).Click()
require.NoError(t, err)
		// one github app installation should be listed
		chromedp.WaitEnabled(`//select[@id='select-install-id']/option[text()='user/leg100']`),
		err := page.Locator(`//button[text()='Create']`).Click()
require.NoError(t, err)
		matchText(t, "//div[@role='alert']", `created provider: github \(app\)`),
	})
}
