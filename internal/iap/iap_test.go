// Package iap contains Google Cloud IAP stuff.
package iap

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
)

func TestIAP_Authenticator(t *testing.T) {
	auth := &Authenticator{
		Audience: "https://example.com",
		Client:   &fakeClient{},
	}

	t.Run("valid iap token", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		r.Header.Add(header, newIAPToken(t, "https://example.com"))
		got, err := auth.Authenticate(w, r)
		require.NoError(t, err)
		assert.Equal(t, &authz.Superuser{}, got)
	})

	t.Run("invalid iap audience", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		r.Header.Add(header, newIAPToken(t, "https://invalid.com"))
		_, err := auth.Authenticate(w, r)
		assert.Error(t, err)
	})
}

func newIAPToken(t *testing.T, aud string) string {
	t.Helper()

	// tests are sometimes run behind an http proxy with a self-signed cert,
	// which the google oauth2 client fails to verify, so just for this test do
	// not use the proxy.
	t.Setenv("HTTPS_PROXY", "")

	credspath := getGoogleCredentialsPath(t)
	src, err := idtoken.NewTokenSource(t.Context(), aud, option.WithAuthCredentialsFile(option.ServiceAccount, credspath))
	require.NoError(t, err)

	token, err := src.Token()
	require.NoError(t, err)
	return token.AccessToken
}

// getGoogleCredentialsPath is a test helper to retrieve the path to a google
// cloud service account key. If the necessary environment variable is not
// present then the test is skipped.
func getGoogleCredentialsPath(t *testing.T) string {
	t.Helper()

	// first try to load the environment variable containing the path to the key
	path, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if !ok {
		// fallback to using an environment variable containing the key itself.
		//
		// NOTE: GOOGLE_CREDENTIALS is set in the github build workflow - if a
		// contributor triggers a PR from a forked repo then GOOGLE_CREDENTIALS
		// is set to an empty string, so skip the test in this scenario.
		key := os.Getenv("GOOGLE_CREDENTIALS")
		if key == "" {
			t.Skip("Export valid GOOGLE_APPLICATION_CREDENTIALS or GOOGLE_CREDENTIALS before running this test")
		}
		path = filepath.Join(t.TempDir(), "google_credentials.json")
		err := os.WriteFile(path, []byte(key), 0o600)
		require.NoError(t, err)
		t.Cleanup(func() {
			os.Remove(path)
		})
	}
	return path
}

type fakeClient struct{}

func (f *fakeClient) GetOrCreateUser(ctx context.Context, username string) (authz.Subject, error) {
	return &authz.Superuser{}, nil
}
