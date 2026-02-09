package tokens

import (
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_StartSession(t *testing.T) {
	key, err := jwk.FromRaw([]byte("abcdef123"))
	require.NoError(t, err)
	svc := Service{
		Logger:       logr.Discard(),
		tokenFactory: &tokenFactory{key: key},
	}

	userID := resource.NewTfeID(resource.UserKind)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/?", nil)
	err = svc.StartSession(w, r, userID)
	require.NoError(t, err)

	// verify and validate token in cookie set in response
	cookies := w.Result().Cookies()
	require.Equal(t, 1, len(cookies))
	token, err := jwt.Parse([]byte(cookies[0].Value), jwt.WithKey(jwa.HS256, key))
	require.NoError(t, err)
	assert.Equal(t, userID.String(), token.Subject())

	// user is redirected to their profile page
	assert.Equal(t, 302, w.Code)
	loc, err := w.Result().Location()
	require.NoError(t, err)
	assert.Equal(t, paths.Profile(), loc.Path)
}
