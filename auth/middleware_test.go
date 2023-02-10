package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AuthenticateToken(t *testing.T) {
	upstream := func(w http.ResponseWriter, r *http.Request) {
		// implicitly respond with 200 OK
	}
	mw := (&authTokenMiddleware{
		UserService:       &fakeUserService{token: "user.token"},
		AgentTokenService: &fakeAgentTokenService{token: "agent.token"},
		siteToken:         "site.token",
	}).handler(http.HandlerFunc(upstream))

	tests := []struct {
		name string
		// add bearer token to http request; nil omits the token
		token *string
		want  int
	}{
		{
			name:  "valid user token",
			token: otf.String("user.token"),
			want:  http.StatusOK,
		},
		{
			name:  "valid site token",
			token: otf.String("site.token"),
			want:  http.StatusOK,
		},
		{
			name:  "valid agent token",
			token: otf.String("agent.token"),
			want:  http.StatusOK,
		},
		{
			name:  "invalid token",
			token: otf.String("invalidToken"),
			want:  http.StatusUnauthorized,
		},
		{
			name:  "malformed token",
			token: otf.String("malfo rmedto ken"),
			want:  http.StatusUnauthorized,
		},
		{
			name: "missing token",
			want: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			if tt.token != nil {
				r.Header.Add("Authorization", "Bearer "+*tt.token)
			}
			mw.ServeHTTP(w, r)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

func Test_AuthenticateUser(t *testing.T) {
	upstream := func(w http.ResponseWriter, r *http.Request) {
		// implicitly respond with 200 OK
	}
	mw := (&authMiddleware{
		Application: &fakeApp{
			fakeUser: otf.NewUser("user-fake"),
		},
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
		assert.Equal(t, paths.Login(), loc.Path)
	})
}
