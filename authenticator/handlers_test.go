package authenticator

import (
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
)

// TestLoginHandler tests the login page handler, testing for the presence of a
// login button for each configured cloud.
func TestLoginHandler(t *testing.T) {
	app := newFakeWebApp(t, nil, withAuthenticators([]*Authenticator{
		{
			oauthClient: &OAuthClient{
				cloudConfig: cloud.Config{Name: "cloud1"},
			},
		},
		{
			oauthClient: &OAuthClient{
				cloudConfig: cloud.Config{Name: "cloud2"},
			},
		},
	}))

	r := httptest.NewRequest("GET", "/?", nil)
	w := httptest.NewRecorder()
	app.loginHandler(w, r)
	body := w.Body.String()
	if assert.Equal(t, 200, w.Code) {
		assert.Contains(t, body, "Login with Cloud1")
		assert.Contains(t, body, "Login with Cloud2")
	}
}
