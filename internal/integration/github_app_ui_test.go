package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_GithubAppNewUI demonstrates creation of a github app via the
// UI.
func TestIntegration_GithubAppNewUI(t *testing.T) {
	integrationTest(t)

	// creating a github app requires site-admin role
	ctx := internal.AddSubjectToContext(context.Background(), &auth.SiteAdmin)

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
			githubHostname := githubStub(t, tt.path, tt.public)
			daemon, _, _ := setup(t, &config{Config: daemon.Config{GithubHostname: githubHostname}})
			tasks := chromedp.Tasks{
				// go to site settings page
				chromedp.Navigate("https://" + daemon.Hostname() + "/app/admin"),
				// go to github app page
				chromedp.Click("//a[text()='GitHub app']"),
				// go to page for creating a new github app
				chromedp.Click("//a[@id='new-github-app-link']"),
			}
			if tt.public {
				tasks = append(tasks, chromedp.Click(`//input[@type='checkbox' and @id='public']`))
			}
			if tt.organization != "" {
				tasks = append(tasks, chromedp.Focus(`//input[@id="organization"]`, chromedp.NodeVisible))
				tasks = append(tasks, input.InsertText(tt.organization))
			}
			tasks = append(tasks, chromedp.Click(`//button[text()='Create']`))
			browser.Run(t, ctx, tasks)
		})
	}

}

func githubStub(t *testing.T, path string, public bool) string {
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		type manifest struct {
			Public bool `json:"public"`
		}
		var params struct {
			Manifest manifest
		}
		require.NoError(t, decode.All(&params, r))
		assert.Equal(t, public, params.Manifest.Public)
	})
	stub := httptest.NewTLSServer(mux)
	t.Cleanup(stub.Close)

	u, err := url.Parse(stub.URL)
	require.NoError(t, err)
	return u.Host
}
