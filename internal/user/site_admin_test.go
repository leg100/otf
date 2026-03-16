package user

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSiteAdmin_Authenticator(t *testing.T) {
	authenticator := &SiteAdminAuthenticator{
		SiteToken: "site-token",
	}
	t.Run("valid site token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		r.Header.Add("Authorization", "Bearer site-token")
		w := httptest.NewRecorder()
		got, err := authenticator.Authenticate(w, r)
		require.NoError(t, err)
		assert.Equal(t, &SiteAdmin, got)
	})

	t.Run("not a site token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		r.Header.Add("Authorization", "Bearer not-a-site-token")
		w := httptest.NewRecorder()
		got, err := authenticator.Authenticate(w, r)
		assert.Nil(t, got)
		assert.Nil(t, err)
	})
}
