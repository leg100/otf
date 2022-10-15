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
	otfhttp "github.com/leg100/otf/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp(t *testing.T) {
	// construct fakes
	fakeOrganization, err := otf.NewOrganization(otf.OrganizationCreateOptions{
		Name: otf.String("org-fake"),
	})
	require.NoError(t, err)

	fakeWorkspace, err := otf.NewWorkspace(fakeOrganization, otf.WorkspaceCreateOptions{
		Name:        "ws-fake",
		LatestRunID: otf.String("run-123"),
	})
	require.NoError(t, err)

	fakeCV, err := otf.NewConfigurationVersion(fakeWorkspace.ID(), otf.ConfigurationVersionCreateOptions{})
	require.NoError(t, err)

	fakeRun := otf.NewRun(fakeCV, fakeWorkspace, otf.RunCreateOptions{})
	require.NoError(t, err)

	fakeUser := otf.NewUser("fake")
	session, err := fakeUser.AttachNewSession(&otf.SessionData{Address: "127.0.0.1"})
	require.NoError(t, err)
	token := session.Token

	fakeAgentToken, err := otf.NewAgentToken(otf.AgentTokenCreateOptions{
		Description:      "fake-token",
		OrganizationName: "org-fake",
	})
	require.NoError(t, err)

	// construct services
	app := &fakeApp{
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
		fakeAgentTokenService: &fakeAgentTokenService{
			fakeAgentToken: fakeAgentToken,
		},
	}
	// Add web app routes.
	router := otfhttp.NewRouter()
	err = AddRoutes(logr.Discard(), &Config{}, app, router)
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
			method: "GET",
			path:   "/organizations/org-fake",
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
			method: "POST",
			path:   "/organizations/org-fake/workspaces/ws-fake/permissions",
			form: url.Values{
				"team_name": []string{"team-123"},
				"role":      []string{"admin"},
			},
			redirect: "/organizations/org-fake/workspaces/ws-fake",
		},
		{
			method: "POST",
			path:   "/organizations/org-fake/workspaces/ws-fake/permissions/unset",
			form: url.Values{
				"team_name": []string{"team-123"},
			},
			redirect: "/organizations/org-fake/workspaces/ws-fake",
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
		{
			method: "GET",
			path:   "/organizations/fake-org/agent-tokens",
		},
		{
			method: "GET",
			path:   "/organizations/fake-org/agent-tokens/new",
		},
		{
			method: "POST",
			path:   "/organizations/fake-org/agent-tokens/create",
			form: url.Values{
				"description": []string{"abcdef"},
			},
			redirect: "/organizations/fake-org/agent-tokens",
		},
		{
			method: "POST",
			path:   "/organizations/fake-org/agent-tokens/delete",
			form: url.Values{
				"id": []string{"ut-fake"},
			},
			redirect: "/organizations/fake-org/agent-tokens",
		},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
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

			t.Log(string(body))

			// check response
			if tt.redirect != "" {
				if assert.Equal(t, 302, res.StatusCode) {
					loc, err := res.Location()
					require.NoError(t, err)
					assert.Equal(t, tt.redirect, loc.Path)
				}
			} else {
				assert.Equal(t, 200, res.StatusCode)
			}
		})
	}
}
