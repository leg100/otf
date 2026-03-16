package user

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSiteAdmin_Authenticator(t *testing.T) {
	authenticator := &SiteAdminAuthenticator{
		SiteToken: "valid-token",
	}
	t.Run("valid site token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		r.Header.Add("Authorization", "Bearer site-token")
		w := httptest.NewRecorder()
		got, err := authenticator.Authenticate(w, r)
		require.NoError(t, err)
		assert.Equal(t, &SiteAdmin, got)
	})

	t.Run("invalid site token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		r.Header.Add("Authorization", "Bearer incorrect")
		w := httptest.NewRecorder()
		_, err := authenticator.Authenticate(w, r)
		assert.Error(t, err)
	})
}
