package user

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSiteAdmin_Authenticator(t *testing.T) {
	t.Run("valid site token", func(t *testing.T) {
		authenticator := &SiteAdminAuthenticator{
			SiteToken: "site-token",
		}
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		r.Header.Add("Authorization", "Bearer site-token")
		w := httptest.NewRecorder()
		got, err := authenticator.Authenticate(w, r)
		require.NoError(t, err)
		assert.Equal(t, &SiteAdmin, got)
	})

	t.Run("invalid site token", func(t *testing.T) {
		authenticator := &SiteAdminAuthenticator{
			SiteToken: "site-token",
		}
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		r.Header.Add("Authorization", "Bearer invalid-site-token")
		w := httptest.NewRecorder()
		got, err := authenticator.Authenticate(w, r)
		assert.Nil(t, got)
		assert.Nil(t, err)
	})

	t.Run("no site token configured, and request has no token set", func(t *testing.T) {
		authenticator := &SiteAdminAuthenticator{}
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		w := httptest.NewRecorder()
		got, err := authenticator.Authenticate(w, r)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
