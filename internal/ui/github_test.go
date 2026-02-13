package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gh "github.com/google/go-github/v65/github"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebHandlers_new(t *testing.T) {
	h := &Handlers{
		HostnameService: internal.NewHostnameService("example.com"),
		GithubHostname:  internal.MustWebURL("github.com"),
	}

	r := httptest.NewRequest("GET", "/?", nil)
	w := httptest.NewRecorder()
	h.newGithubApp(w, r)
	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestWebHandlers_get(t *testing.T) {
	h := &Handlers{
		HostnameService: internal.NewHostnameService("example.com"),
		GithubHostname:  internal.MustWebURL("github.com"),
		GithubApp: &fakeGithubService{
			app: &github.App{},
			installs: []vcs.Installation{
				{ID: 123, Username: new("bob")},
			},
		},
		Authorizer: authz.NewAllowAllAuthorizer(),
	}

	r := httptest.NewRequest("GET", "/?", nil)
	r = r.WithContext(authz.AddSubjectToContext(context.Background(), &user.SiteAdmin))
	w := httptest.NewRecorder()
	h.getGithubApp(w, r)
	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestWebHandlers_exchangeCode(t *testing.T) {
	// create stub github server with an exchange code handler
	stubURL := func() *internal.WebURL {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v3/app-manifests/the-code/conversions", func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal(&gh.AppConfig{
				Slug:  new("my-otf-app"),
				Owner: &gh.User{},
			})
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		stub := httptest.NewTLSServer(mux)
		t.Cleanup(stub.Close)

		u, err := internal.NewWebURL(stub.URL)
		require.NoError(t, err)
		return u
	}()

	h := &Handlers{
		GithubHostname:      stubURL,
		SkipTLSVerification: true,
		GithubApp:           &fakeGithubService{},
	}

	r := httptest.NewRequest("GET", "/?code=the-code", nil)
	w := httptest.NewRecorder()
	h.exchangeCodeGithubApp(w, r)
	testutils.AssertRedirect(t, w, "/app/github-apps")
}

func TestGithubHandlers_deleteApp(t *testing.T) {
	h := &Handlers{
		GithubApp: &fakeGithubService{
			app: &github.App{
				GithubURL: github.DefaultBaseURL(),
			},
		},
	}

	r := httptest.NewRequest("POST", "/?", nil)
	w := httptest.NewRecorder()
	h.deleteGithubApp(w, r)
	testutils.AssertRedirect(t, w, "/app/github-apps")
}

func TestWebHandlers_deleteInstall(t *testing.T) {
	h := &Handlers{
		GithubApp: &fakeGithubService{},
	}

	r := httptest.NewRequest("POST", "/?install_id=123", nil)
	w := httptest.NewRecorder()
	h.deleteGithubAppInstall(w, r)
	testutils.AssertRedirect(t, w, "/app/github-apps")
}

type fakeGithubService struct {
	app      *github.App
	installs []vcs.Installation
}

func (f *fakeGithubService) CreateApp(context.Context, github.CreateAppOptions) (*github.App, error) {
	return f.app, nil
}

func (f *fakeGithubService) GetApp(context.Context) (*github.App, error) {
	return f.app, nil
}

func (f *fakeGithubService) DeleteApp(context.Context) error {
	return nil
}

func (f *fakeGithubService) ListInstallations(context.Context) ([]vcs.Installation, error) {
	return f.installs, nil
}

func (f *fakeGithubService) DeleteInstallation(context.Context, int64) error {
	return nil
}
