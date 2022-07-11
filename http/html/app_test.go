package html

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp(t *testing.T) {
	// setup org
	fakeOrganization, err := otf.NewOrganization(otf.OrganizationCreateOptions{
		Name: otf.String("org-fake"),
	})
	require.NoError(t, err)
	// setup workspace
	fakeWorkspace, err := otf.NewWorkspace(fakeOrganization, otf.WorkspaceCreateOptions{
		Name: "ws-fake",
	})
	require.NoError(t, err)
	// setup configuration version
	fakeCV, err := otf.NewConfigurationVersion(fakeWorkspace.ID(), otf.ConfigurationVersionCreateOptions{})
	require.NoError(t, err)
	// setup run
	fakeRun := otf.NewRun(fakeCV, fakeWorkspace, otf.RunCreateOptions{})
	require.NoError(t, err)
	// setup user
	fakeUser := otf.NewUser("fake")
	session, err := fakeUser.AttachNewSession(&otf.SessionData{Address: "127.0.0.1"})
	require.NoError(t, err)
	token := session.Token
	// setup services
	app := fakeApp{
		fakeUserService: &fakeUserService{
			fakeUser: fakeUser,
		},
		fakeOrganizationService: &fakeOrganizationService{
			fakeOrganization: fakeOrganization,
		},
		fakeWorkspaceService: &fakeWorkspaceService{
			fakeWorkspace: fakeWorkspace,
		},
		fakeRunService: &fakeRunService{
			fakeRun: fakeRun,
		},
	}
	// Add web app routes.
	router := NewRouter()
	err = AddRoutes(logr.Discard(), Config{}, app, router)
	require.NoError(t, err)
	// setup server
	srv := httptest.NewTLSServer(router)
	defer srv.Close()
	// setup client
	client := srv.Client()
	// don't automatically follow redirects
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	tests := []struct {
		path     string
		method   string
		redirect string
		form     url.Values
	}{
		{
			method: "GET",
			path:   "/organizations",
		},
		{
			method:   "GET",
			path:     "/organizations/org-fake",
			redirect: "/organizations/org-fake/workspaces",
		},
		{
			method: "GET",
			path:   "/organizations/org-fake/workspaces",
		},
		{
			method: "GET",
			path:   "/organizations/org-fake/workspaces/new",
		},
		{
			method: "GET",
			path:   "/organizations/org-fake/workspaces/ws-fake/edit",
		},
		{
			method: "POST",
			path:   "/organizations/org-fake/workspaces/ws-fake/update",
			form: url.Values{
				"description": []string{"abcdef"},
			},
			redirect: "/organizations/org-fake/workspaces/ws-fake/edit",
		},
		{
			method:   "POST",
			path:     "/organizations/org-fake/workspaces/create",
			redirect: "/organizations/org-fake/workspaces/ws-fake",
		},
		{
			method: "GET",
			path:   "/organizations/org-fake/workspaces/ws-fake",
		},
		{
			method: "GET",
			path:   "/organizations/org-fake/workspaces/ws-fake/runs",
		},
		{
			method: "GET",
			path:   "/organizations/org-fake/workspaces/ws-fake/runs/" + fakeRun.ID(),
		},
		{
			method: "GET",
			path:   "/profile",
		},
		{
			method: "GET",
			path:   "/profile/sessions",
		},
		{
			method: "GET",
			path:   "/profile/tokens",
		},
		{
			method: "GET",
			path:   "/profile/tokens/new",
		},
		{
			method: "POST",
			path:   "/profile/tokens/create",
			form: url.Values{
				"description": []string{"abcdef"},
			},
			redirect: "/profile/tokens",
		},
		{
			method: "POST",
			path:   "/profile/tokens/delete",
			form: url.Values{
				"id": []string{"ut-fake"},
			},
			redirect: "/profile/tokens",
		},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			// make request
			var reader io.Reader
			if tt.method == "POST" {
				reader = strings.NewReader(tt.form.Encode())
			}
			req, err := http.NewRequest(tt.method, srv.URL+tt.path, reader)
			require.NoError(t, err)
			if tt.method == "POST" {
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			}
			req.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			// check response
			if tt.redirect != "" {
				if assert.Equal(t, 302, res.StatusCode) {
					loc, err := res.Location()
					require.NoError(t, err)
					assert.Equal(t, tt.redirect, loc.Path)
				} else {
					t.Log(string(body))
				}
			} else {
				if !assert.Equal(t, 200, res.StatusCode) {
					t.Log(string(body))
				}
			}
			if res.StatusCode == 500 {
				t.Logf("received http 500; body:\n%s\n", (string(body)))
			}
			//assert.Equal(t, "", string(body))
		})
	}
}
