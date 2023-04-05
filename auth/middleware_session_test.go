package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AuthenticateSession(t *testing.T) {
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// implicitly respond with 200 OK
	})
	mw := AuthenticateSession(&fakeMiddlewareService{
		sessionToken: "session.token",
	})

	t.Run("with session", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/app/organizations", nil)
		r.AddCookie(&http.Cookie{Name: sessionCookie, Value: "session.token"})
		mw(upstream).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
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
