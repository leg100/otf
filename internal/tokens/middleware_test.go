package tokens

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/authz"
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

	t.Run("bearer authenticator", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		w := httptest.NewRecorder()
		r.Header.Add("Authorization", "Bearer valid-token")
		mw := &Middleware{
			authenticators: []authenticator{
				&bearerAuthenticator{
					Client: &fakeBearerAuthenticatorClient{subject: &authz.Superuser{}},
				},
			},
		}
		mw.Authenticate(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
	})
}

var emptyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// implicitly responds with 200 OK
})

type fakeBearerAuthenticatorClient struct {
	subject authz.Subject
}

func (f *fakeBearerAuthenticatorClient) GetSubject(ctx context.Context, token []byte) (authz.Subject, error) {
	return f.subject, nil
}
