package html

import (
	"bytes"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AuthenticateUser(t *testing.T) {
	upstream := func(w http.ResponseWriter, r *http.Request) {
		// implicitly respond with 200 OK
	}
	mw := (&authMiddleware{
		users:    &fakeUserService{fakeUser: otf.NewUser("user-fake")},
		sessions: &fakeSessionService{},
	}).authenticate(http.HandlerFunc(upstream))

	t.Run("with session", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: sessionCookie, Value: "anythingwilldo"})
		mw.ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("without session", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		// deliberately omit session cookie
		mw.ServeHTTP(w, r)
		assert.Equal(t, 302, w.Code)
		loc, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, loginPath(), loc.Path)
	})
}

func Test_SetOrganization(t *testing.T) {
	tests := []struct {
		name string
		// requested path
		path string
		// current organization
		current string
		// wanted organization
		want string
	}{
		{
			name: "new session",
			path: "/non-organization-route",
		},
		{
			name:    "restore session org",
			path:    "/non-organization-route",
			current: "fake-org",
			want:    "fake-org",
		},
		{
			name: "empty session, set org",
			path: "/organizations/fake-org",
			want: "fake-org",
		},
		{
			name:    "same session org",
			path:    "/organizations/fake-org",
			current: "fake-org",
			want:    "fake-org",
		},
		{
			name:    "change session org",
			path:    "/organizations/fake-org",
			current: "previous-org",
			want:    "fake-org",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// handler upstream of middleware
			h := func(w http.ResponseWriter, r *http.Request) {
				// check context contains org
				org, err := organizationFromContext(r.Context())
				if tt.want != "" {
					if assert.NoError(t, err) {
						assert.Equal(t, tt.want, org)
					}
				}
			}
			// setup router and middleware under test
			router := mux.NewRouter()
			router.Use(setOrganization)
			router.HandleFunc("/organizations/{organization_name}", h)
			router.HandleFunc("/non-organization-route", h)
			// setup server
			srv := httptest.NewTLSServer(router)
			defer srv.Close()
			u, err := url.Parse(srv.URL)
			require.NoError(t, err)
			// setup client and its cookie jar
			client := srv.Client()
			jar, err := cookiejar.New(nil)
			require.NoError(t, err)
			if tt.current != "" {
				// populate cookie jar with current session
				jar.SetCookies(u, []*http.Cookie{{Name: organizationCookie, Value: tt.current, Path: "/"}})
			}
			client.Jar = jar
			// make request
			buf := new(bytes.Buffer)
			req, err := http.NewRequest("GET", srv.URL+tt.path, buf)
			require.NoError(t, err)
			_, err = client.Do(req)
			require.NoError(t, err)
			if tt.want != "" {
				// check cookie jar contains wanted session org
				assert.Equal(t, 1, len(client.Jar.Cookies(u)))
				assert.Equal(t, tt.want, client.Jar.Cookies(u)[0].Value)
			}
		})
	}
}
