package session

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
	auth := &Authenticator{
		Client: &fakeClient{},
	}
	t.Run("skip non-protected path", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/login", nil)
		w := httptest.NewRecorder()
		got, err := auth.Authenticate(w, r)
		require.NoError(t, err)
		assert.Equal(t, nil, got)
	})

	t.Run("valid user session", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/app/protected", nil)
		r.AddCookie(&http.Cookie{Name: SessionCookie, Value: ""})
		w := httptest.NewRecorder()
		got, err := auth.Authenticate(w, r)
		require.NoError(t, err)
		assert.Equal(t, &authz.Superuser{}, got)
	})

	t.Run("missing session cookie", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/app/protected", nil)
		w := httptest.NewRecorder()
		got, err := auth.Authenticate(w, r)
		require.NoError(t, err)
		assert.Equal(t, nil, got)
	})
}

type fakeClient struct{}

func (f *fakeClient) GetSubject(ctx context.Context, token []byte) (authz.Subject, error) {
	return &authz.Superuser{}, nil
}
