package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	gogithub "github.com/google/go-github/v65/github"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_VCSProviderUI demonstrates management of personal token vcs providers via
// the UI.
func TestIntegration_VCSProviderTokenUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)

	// create a vcs provider with a github personal access token
	browser.New(t, ctx, func(page playwright.Page) {
		// go to org
		_, err := page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
		require.NoError(t, err)

		screenshot(t, page, "organization_main_menu")
		// go to vcs providers
		err = page.Locator("#menu-item-vcs-providers > a").Click()
		require.NoError(t, err)
		screenshot(t, page, "vcs_providers_list")
		// click 'New Github VCS Provider' button
		err = page.Locator(`//button[text()='New Github-Token Provider']`).Click()
		require.NoError(t, err)
		screenshot(t, page, "new_github_vcs_provider_form")

		// enter fake github token
		err = page.Locator("textarea#token").Fill("fake-github-personal-token")
		require.NoError(t, err)

		// expect default github API URL
		err = expect.Locator(page.Locator(`//input[@name='base_url']`)).ToHaveValue(daemon.GithubHostname.String())
		require.NoError(t, err)

		// submit form to create provider
		err = page.GetByRole("button").Filter(playwright.LocatorFilterOptions{
			HasText: "Create",
		}).Click()
		require.NoError(t, err)

		err = expect.Locator(page.GetByRole("alert")).ToHaveText(`created provider: Github-Token`)
		require.NoError(t, err)

		screenshot(t, page, "vcs_provider_created_github_pat_provider")
		// edit provider
		err = page.Locator(`//button[@id='edit-button']`).Click()
		require.NoError(t, err)

		// give it a name
		err = page.Locator("input#name").Fill("my-token")
		require.NoError(t, err)

		err = page.Locator(`//button[text()='Update']`).Click()
		require.NoError(t, err)
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("updated provider: my-token")
		require.NoError(t, err)

		// change token
		err = page.Locator(`//button[@id='edit-button']`).Click()
		require.NoError(t, err)

		err = page.Locator("textarea#token").Fill("my-updated-fake-github-personal-token")
		require.NoError(t, err)

		err = page.Locator(`//button[text()='Update']`).Click()
		require.NoError(t, err)

		err = expect.Locator(page.GetByRole("alert")).ToHaveText("updated provider: my-token")
		require.NoError(t, err)

		// clear name
		err = page.Locator(`//button[@id='edit-button']`).Click()
		require.NoError(t, err)

		err = page.Locator("input#name").Clear()
		require.NoError(t, err)

		err = page.Locator(`//button[text()='Update']`).Click()
		require.NoError(t, err)

		err = expect.Locator(page.GetByRole("alert")).ToHaveText(`updated provider: Github-Token`)
		require.NoError(t, err)

		// change API URL
		err = page.Locator(`//button[@id='edit-button']`).Click()
		require.NoError(t, err)

		err = page.Locator(`//input[@name='base_url']`).Fill("http://my-overpriced-github-enterprise-server/api")
		require.NoError(t, err)

		err = page.Locator(`//button[text()='Update']`).Click()
		require.NoError(t, err)

		err = expect.Locator(page.GetByRole("alert")).ToHaveText(`updated provider: Github-Token`)
		require.NoError(t, err)

		err = page.Locator(`//button[@id='edit-button']`).Click()
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//input[@name='base_url']`)).ToHaveValue(`http://my-overpriced-github-enterprise-server/api`)
		require.NoError(t, err)

		// delete token
		err = page.Locator(`//button[@id='delete-vcs-provider-button']`).Click()
		require.NoError(t, err)

		err = expect.Locator(page.GetByRole("alert")).ToHaveText(`deleted provider: Github-Token`)
		require.NoError(t, err)
	})
}

// TestIntegration_VCSProviderAppUI demonstrates management of github app vcs
// providers via the UI.
func TestIntegration_VCSProviderAppUI(t *testing.T) {
	integrationTest(t)

	// create github stub server and return its hostname.
	githubHostname := func(t *testing.T) string {
		install := &gogithub.Installation{
			ID:    internal.Ptr[int64](123),
			AppID: internal.Ptr[int64](456),
			Account: &gogithub.User{
				Login: internal.Ptr("leg100"),
				Type:  internal.Ptr("User"),
			},
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
				Repositories: []*gogithub.Repository{{FullName: internal.Ptr("leg100/otf-workspaces")}},
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

	daemon, org, _ := setup(t, withGithubHostname(githubHostname))

	// creating a github app requires site-admin role
	ctx := authz.AddSubjectToContext(context.Background(), &user.SiteAdmin)

	// create app
	_, err := daemon.GithubApp.CreateApp(ctx, github.CreateAppOptions{
		BaseURL:    daemon.GithubHostname,
		AppID:      456,
		Slug:       "otf-123",
		PrivateKey: string(testutils.ReadFile(t, "./fixtures/key.pem")),
	})
	require.NoError(t, err)

	// create github app vcs provider via UI.
	browser.New(t, ctx, func(page playwright.Page) {
		// go to org
		_, err = page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
		require.NoError(t, err)

		// go to vcs providers
		err = page.Locator("#menu-item-vcs-providers > a").Click()
		require.NoError(t, err)

		screenshot(t, page, "vcs_provider_list_including_github_app")

		// click button for creating a new vcs provider with a github app
		err = page.GetByRole("button").Filter(playwright.LocatorFilterOptions{
			HasText: "New Github-App Provider",
		}).Click()
		require.NoError(t, err)

		// one github app installation should be listed
		err = expect.Locator(page.Locator(`//select[@id='select-install-id']/option[text()='user/leg100']`)).ToBeAttached()
		require.NoError(t, err)

		err = page.GetByRole("button").Filter(playwright.LocatorFilterOptions{HasText: "Create"}).Click()
		require.NoError(t, err)

		err = expect.Locator(page.GetByRole("alert")).ToHaveText(`created provider: Github-App`)
		require.NoError(t, err)
	})
}
