package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
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
