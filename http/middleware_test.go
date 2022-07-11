package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AuthMiddleware(t *testing.T) {
	upstream := func(w http.ResponseWriter, r *http.Request) {
		// implicitly respond with 200 OK
	}
	mw := (&authTokenMiddleware{
		svc:       &fakeUserService{token: "validUserToken"},
		siteToken: "validSiteToken",
	}).handler(http.HandlerFunc(upstream))

	tests := []struct {
		name  string
		token string
		want  int
	}{
		{
			name:  "valid user token",
			token: "validUserToken",
			want:  http.StatusOK,
		},
		{
			name:  "valid site token",
			token: "validSiteToken",
			want:  http.StatusOK,
		},
		{
			name:  "invalid token",
			token: "invalidToken",
			want:  http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			r.Header.Add("Authorization", "Bearer "+tt.token)
			mw.ServeHTTP(w, r)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}
