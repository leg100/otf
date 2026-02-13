package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	gogithub "github.com/google/go-github/v65/github"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_GithubAppsUI tests management of github apps via the UI.
func TestIntegration_GithubAppsUI(t *testing.T) {
	integrationTest(t)

	// creating a github app requires site-admin role
	ctx := authz.AddSubjectToContext(context.Background(), &user.SiteAdmin)

	// these tests submit the create github app form using different
	// combinations of form fields, and then checking that a (stub) github server
	// receives the completed form correctly.
	tests := []struct {
		name         string
		public       bool   // whether to tick 'public' checkbox
		organization string // install in organization github account
		path         string // form should submitted to this path on github
	}{
		{
			"create private app in personal github account",
			false,
			"",
			"/settings/apps/new",
		},
		{
			"create public app in personal github account",
			true,
			"",
			"/settings/apps/new",
		},
		{
			"create private app in organization github account",
			false,
			"acme-corp",
			"/organizations/acme-corp/settings/apps/new",
		},
		{
			"create public app in organization github account",
			true,
			"acme-corp",
			"/organizations/acme-corp/settings/apps/new",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemon, _, _ := setup(t, withGithubOption(github.WithHandler(tt.path, func(w http.ResponseWriter, r *http.Request) {
				// check that the manifest has been correctly submitted.
				var (
					manifest struct {
						Public bool
					}
					params struct {
						Manifest string `schema:"manifest,required"`
					}
				)
				// first decode POST form manifest=<json>
				err := decode.Form(&params, r)
				require.NoError(t, err)
				// then unmarshal the json into a manifest
				err = json.Unmarshal([]byte(params.Manifest), &manifest)
				require.NoError(t, err)
				require.Equal(t, tt.public, manifest.Public)
				w.Write([]byte(`<html><body>success</body></html>`))
			})))

			browser.New(t, ctx, func(page playwright.Page) {
				// go to site settings page
				_, err := page.Goto("https://" + daemon.System.Hostname() + "/app/admin")
				require.NoError(t, err)
				screenshot(t, page, "site_settings")

				// go to github app page
				err = page.Locator("#menu-item-github-app > a").Click()
				require.NoError(t, err)

				screenshot(t, page, "empty_github_app_page")

				// go to page for creating a new github app
				err = page.Locator("//a[@id='new-github-app-link']").Click()
				require.NoError(t, err)

				screenshot(t, page, "new_github_app")

				if tt.public {
					err = page.Locator(`//input[@type='checkbox' and @id='public']`).Click()
					require.NoError(t, err)
				}

				if tt.organization != "" {
					err = page.Locator(`//input[@id="organization"]`).Fill(tt.organization)
					require.NoError(t, err)
				}

				err = page.GetByRole("button").Filter(playwright.LocatorFilterOptions{
					HasText: "Create",
				}).Click()
				require.NoError(t, err)

				err = expect.Locator(page.GetByText("success")).ToBeVisible()
				require.NoError(t, err)

			})
		})
	}

	// demonstrate the completion of creating a github app, by taking over from
	// where Github would redirect back to OTF, exchanging the code with a
	// stub Github server, and receiving back the app config, and then
	// redirecting to the github app page showing the created app.
	t.Run("complete creation of github app", func(t *testing.T) {
		daemon, _, _ := setup(t,
			withGithubOption(github.WithHandler("/api/v3/app-manifests/anything/conversions", func(w http.ResponseWriter, r *http.Request) {
				out, err := json.Marshal(&gogithub.AppConfig{
					ID:            new(int64(123)),
					Slug:          new("my-otf-app"),
					WebhookSecret: new("top-secret"),
					PEM:           new(string(testutils.ReadFile(t, "./fixtures/key.pem"))),
					Owner:         &gogithub.User{},
				})
				require.NoError(t, err)
				w.Header().Add("Content-Type", "application/json")
				w.Write(out)
			})),
			withGithubOption(github.WithHandler("/api/v3/app/installations", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
			})),
		)
		browser.New(t, ctx, func(page playwright.Page) {
			// go to the exchange code endpoint
			_, err := page.Goto((&url.URL{
				Scheme:   "https",
				Host:     daemon.System.Hostname(),
				Path:     "/app/github-apps/exchange-code",
				RawQuery: "code=anything",
			}).String())
			require.NoError(t, err)

			err = expect.Locator(page.Locator(`//tr[@id='item-github-app']//td[1]`)).ToHaveText("my-otf-app")
			require.NoError(t, err)

			screenshot(t, page, "github_app_created")

			err = expect.Locator(page.GetByRole("alert")).ToHaveText(`created github app: my-otf-app`)
			require.NoError(t, err)
		})
	})

	// demonstrate the listing of github installations
	t.Run("list github app installs", func(t *testing.T) {
		handler := github.WithHandler("/api/v3/app/installations", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			out, err := json.Marshal([]*gogithub.Installation{
				{
					ID:    new(int64(123)),
					AppID: new(int64(123)),
					Account: &gogithub.User{
						Login: new("leg100"),
						Type:  new("User"),
					},
				},
			})
			require.NoError(t, err)
			w.Write(out)
		})
		daemon, _, _ := setup(t, withGithubOption(handler))
		_, err := daemon.GithubApp.CreateApp(ctx, github.CreateAppOptions{
			BaseURL:    daemon.GithubHostname,
			AppID:      123,
			Slug:       "otf-123",
			PrivateKey: string(testutils.ReadFile(t, "./fixtures/key.pem")),
		})
		require.NoError(t, err)

		browser.New(t, ctx, func(page playwright.Page) {
			_, err = page.Goto(daemon.System.URL("/app/github-apps"))
			require.NoError(t, err)

			err = expect.Locator(page.Locator(`//div[@id='installations']//tbody//td[1]//a`)).ToContainText("user/leg100")
			require.NoError(t, err)

			screenshot(t, page, "github_app_install_list")
		})
	})

	// demonstrate removing a github app via the UI
	t.Run("delete github app", func(t *testing.T) {
		handler := github.WithHandler("/api/v3/app/installations", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			out, err := json.Marshal([]*gogithub.Installation{
				{
					ID:    new(int64(123)),
					AppID: new(int64(123)),
					Account: &gogithub.User{
						Login: new("leg100"),
						Type:  new("User"),
					},
				},
			})
			require.NoError(t, err)
			w.Write(out)
		})
		daemon, _, _ := setup(t, withGithubOption(handler))

		_, err := daemon.GithubApp.CreateApp(ctx, github.CreateAppOptions{
			AppID:      123,
			Slug:       "my-otf-app",
			PrivateKey: string(testutils.ReadFile(t, "./fixtures/key.pem")),
			BaseURL:    daemon.GithubHostname,
		})
		require.NoError(t, err)

		browser.New(t, ctx, func(page playwright.Page) {
			_, err = page.Goto(daemon.System.URL("/app/github-apps"))
			require.NoError(t, err)

			err = expect.Locator(page.Locator(`//tr[@id='item-github-app']//td[1]`)).ToHaveText("my-otf-app")
			require.NoError(t, err)

			err = page.Locator(`//tr[@id='item-github-app']//button[@id='delete-button']`).Click()
			require.NoError(t, err)

			err = expect.Locator(page.GetByRole("alert")).ToHaveText(`Deleted GitHub app my-otf-app from OTF. You still need to delete the app in GitHub.`)
			require.NoError(t, err)
		})
	})
}

// TestIntegration_GithubApp_Event demonstrates an event from a github
// app installation triggering a run.
func TestIntegration_GithubApp_Event(t *testing.T) {
	integrationTest(t)

	// create an OTF daemon with a fake github backend, and serve up a repo and
	// its contents via tarball.
	daemon, org, ctx := setup(t, withGithubOptions(
		github.WithRepo(vcs.NewMustRepo("leg100", "otf-workspaces")),
		github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
		github.WithHandler("/api/v3/app/installations/42997659", func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal(&gogithub.Installation{
				ID:    new(int64(42997659)),
				AppID: new(int64(123)),
				Account: &gogithub.User{
					Login: new("leg100"),
					Type:  new("User"),
				},
			})
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		}),
		github.WithHandler("/api/v3/app/installations/42997659/access_tokens", func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal(&gogithub.InstallationToken{})
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		}),
	))
	// creating a github app requires site-admin role
	ctx = authz.AddSubjectToContext(ctx, &user.SiteAdmin)
	_, err := daemon.GithubApp.CreateApp(ctx, github.CreateAppOptions{
		BaseURL: daemon.GithubHostname,
		AppID:   123,
		// any key will do, the stub github server won't actually authenticate it.
		PrivateKey:    string(testutils.ReadFile(t, "./fixtures/key.pem")),
		Slug:          "test-app",
		WebhookSecret: "secret",
	})
	require.NoError(t, err)

	provider, err := daemon.VCSProviders.Create(ctx, vcs.CreateOptions{
		Organization: org.Name,
		KindID:       github.AppKindID,
		InstallID:    new(int64(42997659)),
	})
	require.NoError(t, err)

	// create and connect a workspace to a repo using the app install
	_, err = daemon.Workspaces.Create(ctx, workspace.CreateOptions{
		Name:         new("dev"),
		Organization: &org.Name,
		ConnectOptions: &workspace.ConnectOptions{
			VCSProviderID: &provider.ID,
			RepoPath:      new(vcs.NewMustRepo("leg100", "otf-workspaces")),
		},
	})
	require.NoError(t, err)

	// send event
	push := testutils.ReadFile(t, "./fixtures/github_app_push.json")
	github.SendEventRequest(t, github.PushEvent, daemon.System.URL(github.AppEventsPath), "secret", push)

	runEvent := <-daemon.runEvents
	daemon.waitRunStatus(t, ctx, runEvent.Payload.ID, runstatus.Planned)

	// github should receive a pending status update
	got := daemon.GetStatus(t, ctx)
	assert.Equal(t, "pending", got.GetState())
}
