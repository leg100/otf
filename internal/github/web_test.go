package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v65/github"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebHandlers_new(t *testing.T) {
	h := &webHandlers{
		HostnameService: internal.NewHostnameService("example.com"),
		githubAPIURL:    internal.MustWebURL("github.com"),
	}

	r := httptest.NewRequest("GET", "/?", nil)
	w := httptest.NewRecorder()
	h.new(w, r)
	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestWebHandlers_get(t *testing.T) {
	h := &webHandlers{
		HostnameService: internal.NewHostnameService("example.com"),
		githubAPIURL:    internal.MustWebURL("github.com"),
		svc: &fakeService{
			app: &App{},
			installs: []vcs.Installation{
				{ID: 123, Username: internal.Ptr("bob")},
			},
		},
	}

	r := httptest.NewRequest("GET", "/?", nil)
	r = r.WithContext(authz.AddSubjectToContext(context.Background(), &user.SiteAdmin))
	w := httptest.NewRecorder()
	h.get(w, r)
	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestWebHandlers_exchangeCode(t *testing.T) {
	// create stub github server with an exchange code handler
	stubURL := func() *internal.WebURL {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v3/app-manifests/the-code/conversions", func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal(&github.AppConfig{
				Slug:  internal.Ptr("my-otf-app"),
				Owner: &github.User{},
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

	h := &webHandlers{
		githubAPIURL:        stubURL,
		skipTLSVerification: true,
		svc:                 &fakeService{},
	}

	r := httptest.NewRequest("GET", "/?code=the-code", nil)
	w := httptest.NewRecorder()
	h.exchangeCode(w, r)
	testutils.AssertRedirect(t, w, "/app/github-apps")
}

func TestWebHandlers_deleteApp(t *testing.T) {
	h := &webHandlers{
		svc: &fakeService{
			app: &App{
				GithubURL: DefaultBaseURL,
			},
		},
	}

	r := httptest.NewRequest("POST", "/?", nil)
	w := httptest.NewRecorder()
	h.delete(w, r)
	testutils.AssertRedirect(t, w, "/app/github-apps")
}

func TestWebHandlers_deleteInstall(t *testing.T) {
	h := &webHandlers{
		svc: &fakeService{},
	}

	r := httptest.NewRequest("POST", "/?install_id=123", nil)
	w := httptest.NewRecorder()
	h.deleteInstall(w, r)
	testutils.AssertRedirect(t, w, "/app/github-apps")
}

type fakeService struct {
	app      *App
	installs []vcs.Installation
}

func (f *fakeService) CreateApp(context.Context, CreateAppOptions) (*App, error) {
	return f.app, nil
}

func (f *fakeService) GetApp(context.Context) (*App, error) {
	return f.app, nil
}

func (f *fakeService) DeleteApp(context.Context) error {
	return nil
}

func (f *fakeService) ListInstallations(context.Context) ([]vcs.Installation, error) {
	return f.installs, nil
}

func (f *fakeService) DeleteInstallation(context.Context, int64) error {
	return nil
}
