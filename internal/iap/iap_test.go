// Package iap contains Google Cloud IAP stuff.
package iap

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/idtoken"
)

func TestIAP_Authenticator(t *testing.T) {
	auth := &Authenticator{
		audience: "https://example.com",
		users:    &fakeClient{subject: &authz.Superuser{}},
		validator: tokenValidatorFunc(func(ctx context.Context, s1, s2 string) (*idtoken.Payload, error) {
			return &idtoken.Payload{
				Claims: map[string]any{
					"email": "user@example.com",
				},
			}, nil
		}),
	}

	t.Run("is an iap request", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		r.Header.Add(header, "google-jwt")
		got, err := auth.Authenticate(w, r)
		require.NoError(t, err)
		assert.Equal(t, &authz.Superuser{}, got)
	})

	t.Run("not an iap request", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		// no iap header
		got, err := auth.Authenticate(w, r)
		assert.Nil(t, got)
		assert.Nil(t, err)
	})
}

type fakeClient struct {
	subject authz.Subject
}

func (f *fakeClient) GetOrCreateUser(ctx context.Context, username string) (authz.Subject, error) {
	return f.subject, nil
}
