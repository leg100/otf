package authenticator

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestOAuthAuthenticator_ResponseHandler(t *testing.T) {
	user := cloud.User{Name: "fake-user"}

	authenticator := &oauthAuthenticator{
		TokensService: &fakeAuthenticatorService{},
		oauthClient: &fakeOAuthClient{
			user:  user,
			token: &oauth2.Token{},
		},
	}

	r := httptest.NewRequest("GET", "/auth?state=state", nil)
	r.AddCookie(&http.Cookie{Name: oauthCookieName, Value: "state"})
	w := httptest.NewRecorder()
	authenticator.ResponseHandler(w, r)

	assert.Equal(t, http.StatusFound, w.Result().StatusCode)

	loc, err := w.Result().Location()
	require.NoError(t, err)
	assert.Equal(t, "/app/profile", loc.Path)
}
