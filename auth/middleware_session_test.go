package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/leg100/otf/http/html/paths"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AuthenticateSession(t *testing.T) {
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// implicitly respond with 200 OK
	})
	secret := "abcdef123"
	mw, err := NewAuthSessionMiddleware(&fakeMiddlewareService{
		sessionToken: "session.token",
	}, secret)
	require.NoError(t, err)

	t.Run("with session", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/app/organizations", nil)
		token := newTestJWT(t, secret, time.Now().Add(time.Minute))
		r.AddCookie(&http.Cookie{Name: sessionCookie, Value: string(token)})
		mw(upstream).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("with expired session", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/app/organizations", nil)
		token := newTestJWT(t, secret, time.Now().Add(-time.Minute))
		r.AddCookie(&http.Cookie{Name: sessionCookie, Value: string(token)})
		mw(upstream).ServeHTTP(w, r)

		assert.Equal(t, 302, w.Code)
		loc, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, paths.Login(), loc.Path)
	})

	t.Run("without session", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/app/organizations", nil)
		// deliberately omit session cookie
		mw(upstream).ServeHTTP(w, r)

		assert.Equal(t, 302, w.Code)
		loc, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, paths.Login(), loc.Path)
	})
}

func newTestJWT(t *testing.T, key string, expiry time.Time) []byte {
	t.Helper()

	token, err := jwt.NewBuilder().
		IssuedAt(time.Now()).
		Expiration(expiry).
		Build()
	require.NoError(t, err)
	serialized, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, []byte(key)))
	require.NoError(t, err)
	return serialized
}
