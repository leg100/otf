package html

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
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
	// setup ws
	fakeWorkspace, err := otf.NewWorkspace(fakeOrganization.ID(), otf.WorkspaceCreateOptions{
		Name: "ws-fake",
	})
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
	}
	// Add web app routes.
	router := mux.NewRouter()
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
		redirect string
	}{
		{
			path: "/organizations",
		},
		{
			path:     "/organizations/org-fake",
			redirect: "/organizations/org-fake/workspaces",
		},
		{
			path: "/organizations/org-fake/workspaces",
		},
		{
			path: "/organizations/org-fake/workspaces/ws-fake",
		},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			buf := new(bytes.Buffer)
			req, err := http.NewRequest("GET", srv.URL+tt.path, buf)
			require.NoError(t, err)
			req.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()
			if tt.redirect != "" {
				assert.Equal(t, 302, res.StatusCode)
				loc, err := res.Location()
				require.NoError(t, err)
				assert.Equal(t, tt.redirect, loc.Path)
			} else {
				assert.Equal(t, 200, res.StatusCode)
			}
			if res.StatusCode == 500 {
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				t.Logf("received http 500; body:\n%s\n", (string(body)))
			}
			//assert.Equal(t, "", string(body))
		})
	}
}
