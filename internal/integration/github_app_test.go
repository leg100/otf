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
	gogithub "github.com/google/go-github/v55/github"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

// TestIntegration_GithubAppNewUI demonstrates creation of a github app via the
// UI.
func TestIntegration_GithubAppNewUI(t *testing.T) {
	integrationTest(t)

	// creating a github app requires site-admin role
	ctx := internal.AddSubjectToContext(context.Background(), &user.SiteAdmin)

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
			githubHostname := func(t *testing.T, path string, public bool) string {
				mux := http.NewServeMux()
				mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
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
					require.Equal(t, public, manifest.Public)
					w.Write([]byte(`<html><body>success</body></html>`))
				})
				mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					t.Fatalf("form submitted to wrong path: %s", r.URL.Path)
				})
				stub := httptest.NewTLSServer(mux)
				t.Cleanup(stub.Close)

				u, err := url.Parse(stub.URL)
				require.NoError(t, err)
				return u.Host
			}(t, tt.path, tt.public)

			daemon, _, _ := setup(t, &config{Config: daemon.Config{GithubHostname: githubHostname}})
			tasks := chromedp.Tasks{
				// go to site settings page
				chromedp.Navigate("https://" + daemon.Hostname() + "/app/admin"),
				screenshot(t, "site_settings"),
				// go to github app page
				chromedp.Click("//a[text()='GitHub app']"),
				screenshot(t, "empty_github_app_page"),
				// go to page for creating a new github app
				chromedp.Click("//a[@id='new-github-app-link']"),
				screenshot(t, "new_github_app"),
			}
			if tt.public {
				tasks = append(tasks, chromedp.Click(`//input[@type='checkbox' and @id='public']`))
			}
			if tt.organization != "" {
				tasks = append(tasks, chromedp.Focus(`//input[@id="organization"]`, chromedp.NodeVisible))
				tasks = append(tasks, input.InsertText(tt.organization))
			}
			tasks = append(tasks, chromedp.Click(`//button[text()='Create']`))
			tasks = append(tasks, chromedp.WaitVisible(`//body[text()='success']`))
			browser.Run(t, ctx, tasks)
		})
	}

	// demonstrate the completion of creating a github app, by taking over from
	// where Github would redirect back to OTF, exchanging the code with a
	// stub Github server, and receiving back the app config, and then
	// redirecting to the github app page showing the created app.
	t.Run("complete creation of github app", func(t *testing.T) {
		handlers := []github.TestServerOption{
			github.WithHandler("/api/v3/app-manifests/anything/conversions", func(w http.ResponseWriter, r *http.Request) {
				out, err := json.Marshal(&gogithub.AppConfig{
					ID:            internal.Int64(123),
					Slug:          internal.String("my-otf-app"),
					WebhookSecret: internal.String("top-secret"),
					PEM:           internal.String(string(testutils.ReadFile(t, "./fixtures/key.pem"))),
					Owner:         &gogithub.User{},
				})
				require.NoError(t, err)
				w.Header().Add("Content-Type", "application/json")
				w.Write(out)
			}),
			github.WithHandler("/api/v3/app/installations", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
			}),
		}
		daemon, _, _ := setup(t, nil, handlers...)
		browser.Run(t, ctx, chromedp.Tasks{
			// go to the exchange code endpoint
			chromedp.Navigate((&url.URL{
				Scheme:   "https",
				Host:     daemon.Hostname(),
				Path:     "/app/github-apps/exchange-code",
				RawQuery: "code=anything",
			}).String()),
			chromedp.WaitVisible(`//div[@class='widget']//a[contains(text(), "my-otf-app")]`),
			screenshot(t, "github_app_created"),
		})
	})

	// demonstrate the listing of github installations
	t.Run("list github app installs", func(t *testing.T) {
		handlers := []github.TestServerOption{
			github.WithHandler("/api/v3/app/installations", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				out, err := json.Marshal([]*gogithub.Installation{
					{
						ID:      internal.Int64(123),
						Account: &gogithub.User{Login: internal.String("leg100")},
					},
				})
				require.NoError(t, err)
				w.Header().Add("Content-Type", "application/json")
				w.Write(out)
			}),
		}
		daemon, _, _ := setup(t, nil, handlers...)
		_, err := daemon.CreateGithubApp(ctx, github.CreateAppOptions{
			AppID:      123,
			Slug:       "otf-123",
			PrivateKey: string(testutils.ReadFile(t, "./fixtures/key.pem")),
		})
		require.NoError(t, err)
		browser.Run(t, ctx, chromedp.Tasks{
			chromedp.Navigate(daemon.HostnameService.URL("/app/github-apps")),
			chromedp.WaitVisible(`//div[@id='installations']//a[contains(text(), "user/leg100")]`),
			screenshot(t, "github_app_install_list"),
		})
	})
}

// TestIntegration_GithubApp_Event demonstrates an event from a github
// app installation triggering a run.
func TestIntegration_GithubApp_Event(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil,
		github.WithRepo("leg100/otf-workspaces"),
		github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
		github.WithHandler("/api/v3/app/installations/42997659", func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal(&gogithub.Installation{
				ID:         internal.Int64(42997659),
				Account:    &gogithub.User{Login: internal.String("leg100")},
				TargetType: internal.String("User"),
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
	)
	// creating a github app requires site-admin role
	ctx = internal.AddSubjectToContext(ctx, &user.SiteAdmin)
	// create an OTF daemon with a fake github backend, and serve up a repo and
	// its contents via tarball.
	_, err := daemon.CreateGithubApp(ctx, github.CreateAppOptions{
		// any key will do, the stub github server won't actually authenticate it.
		PrivateKey:    string(testutils.ReadFile(t, "./fixtures/key.pem")),
		Slug:          "test-app",
		WebhookSecret: "secret",
	})
	require.NoError(t, err)

	provider, err := daemon.CreateVCSProvider(ctx, vcsprovider.CreateOptions{
		Organization:       org.Name,
		GithubAppInstallID: internal.Int64(42997659),
	})
	require.NoError(t, err)

	// create and connect a workspace to a repo using the app install
	_, err = daemon.Workspaces.CreateWorkspace(ctx, workspace.CreateOptions{
		Name:         internal.String("dev"),
		Organization: internal.String(org.Name),
		ConnectOptions: &workspace.ConnectOptions{
			VCSProviderID: &provider.ID,
			RepoPath:      internal.String("leg100/otf-workspaces"),
		},
	})
	require.NoError(t, err)

	sub, unsub := daemon.Runs.WatchRuns(ctx)
	defer unsub()

	// send event
	push := testutils.ReadFile(t, "./fixtures/github_app_push.json")
	github.SendEventRequest(t, github.PushEvent, daemon.HostnameService.URL(github.AppEventsPath), "secret", push)

	// wait for run to be created
	<-sub
}
