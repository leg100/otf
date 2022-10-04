package html

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	oauth2gitlab "golang.org/x/oauth2/gitlab"
)

func TestAuthenticator_RequestHandler(t *testing.T) {
	authenticator := NewAuthenticator(&fakeAuthenticatorApp{}, &fakeCloud{})

	r := httptest.NewRequest("GET", "/auth", nil)
	w := httptest.NewRecorder()
	authenticator.requestHandler(w, r)

	assert.Equal(t, http.StatusFound, w.Result().StatusCode)

	loc, err := w.Result().Location()
	require.NoError(t, err)
	assert.Equal(t, "gitlab.com", loc.Host)
	assert.Equal(t, "http://example.com/oauth/fake/callback", loc.Query().Get("redirect_uri"))

	if assert.Equal(t, 1, len(w.Result().Cookies())) {
		assert.Equal(t, w.Result().Cookies()[0].Value, loc.Query().Get("state"))
	}
}

func TestAuthenticator_ResponseHandler(t *testing.T) {
	authenticator := NewAuthenticator(&fakeAuthenticatorApp{}, &fakeCloud{})

	r := httptest.NewRequest("GET", "/auth", nil)
	w := httptest.NewRecorder()
	authenticator.responseHandler(w, r)

	assert.Equal(t, http.StatusFound, w.Result().StatusCode)

	loc, err := w.Result().Location()
	require.NoError(t, err)
	assert.Equal(t, "/login", loc.Path)

	assert.Equal(t, 1, len(w.Result().Cookies()))
}

type fakeAuthenticatorApp struct {
	otf.Application
}

type fakeCloud struct {
	*OAuthCredentials
}

func (f *fakeCloud) CloudName() string         { return "fake" }
func (f *fakeCloud) Endpoint() oauth2.Endpoint { return oauth2gitlab.Endpoint }
func (f *fakeCloud) Scopes() []string          { return []string{} }
func (f *fakeCloud) ClientID() string          { return "abc123" }
func (f *fakeCloud) ClientSecret() string      { return "xyz789" }
func (f *fakeCloud) NewDirectoryClient(context.Context, DirectoryClientOptions) (DirectoryClient, error) {
	return nil, nil
}
