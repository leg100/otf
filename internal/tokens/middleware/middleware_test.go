package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	t.Run("skip non-protected path", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/login", nil)
		w := httptest.NewRecorder()
		mw := &Middleware{}
		mw.Authenticate(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)

		// Should be no subject in context.
		_, err := authz.SubjectFromContext(t.Context())
		require.Error(t, err)
	})

	t.Run("no authenticators", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		w := httptest.NewRecorder()
		mw := &Middleware{}
		mw.Authenticate(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 401, w.Code)
	})

	t.Run("allow request to protected api path", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		w := httptest.NewRecorder()
		mw := &Middleware{
			APIAuthenticators: []Authenticator{
				&fakeAuthenticator{subject: &authz.Superuser{}},
			},
		}
		mw.Authenticate(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("disallow request to protected api path", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		w := httptest.NewRecorder()
		mw := &Middleware{
			APIAuthenticators: []Authenticator{
				&fakeAuthenticator{subject: nil},
			},
		}
		mw.Authenticate(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 401, w.Code)
	})

	t.Run("allow request to protected UI path", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/app/protected", nil)
		w := httptest.NewRecorder()
		mw := &Middleware{
			UIAuthenticators: []Authenticator{
				&fakeAuthenticator{subject: &authz.Superuser{}},
			},
		}
		mw.Authenticate(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("disallow request to protected UI path", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/app/protected", nil)
		w := httptest.NewRecorder()
		mw := &Middleware{
			UIAuthenticators: []Authenticator{
				&fakeAuthenticator{subject: nil},
			},
		}
		mw.Authenticate(emptyHandler).ServeHTTP(w, r)

		// Should redirect to login page
		assert.Equal(t, 302, w.Code)
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, path.Login(), redirect.Path)
	})
}

var emptyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// implicitly responds with 200 OK
})

type fakeAuthenticator struct {
	subject authz.Subject
}

func (f *fakeAuthenticator) Authenticate(w http.ResponseWriter, r *http.Request) (authz.Subject, error) {
	return f.subject, nil
}
