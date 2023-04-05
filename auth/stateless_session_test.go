package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html/paths"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatelessSessionService(t *testing.T) {
	secret := "abcdef123"
	svc, err := newStatelessSessionService(logr.Discard(), secret)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/?", nil)
	svc.StartSession(w, r, CreateStatelessSessionOptions{
		Username: otf.String("bobby"),
	})

	// verify and validate token in cookie set in response
	cookies := w.Result().Cookies()
	require.Equal(t, 1, len(cookies))
	token, err := jwt.Parse([]byte(cookies[0].Value), jwt.WithKey(jwa.HS256, []byte(secret)))
	require.NoError(t, err)
	assert.Equal(t, "bobby", token.Subject())

	// user is redirected to their profile page
	assert.Equal(t, 302, w.Code)
	loc, err := w.Result().Location()
	require.NoError(t, err)
	assert.Equal(t, paths.Profile(), loc.Path)
}
