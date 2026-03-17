package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/iap"
	"github.com/leg100/otf/internal/organization"
	userpkg "github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
)

// TestAuthenticators tests the various methods of remotely authenticating with
// the system.
func TestAuthenticators(t *testing.T) {
	integrationTest(t)

	// Configure site token for the site admin token test below.
	siteToken := "site-token"
	daemon, org, ctx := setup(t, withSiteToken(siteToken))

	// Retrieve user for the user token test below.
	user, err := userpkg.UserFromContext(ctx)
	require.NoError(t, err)

	// Create org token for the org token test below.
	ot, orgToken, err := daemon.Organizations.CreateOrganizationToken(
		ctx,
		organization.CreateOrganizationTokenOptions{Organization: org.Name},
	)
	require.NoError(t, err)

	tests := []struct {
		name        string
		setup       func(t *testing.T, r *http.Request)
		wantCode    int
		wantSubject any
	}{
		{
			name:     "unauthenticated",
			wantCode: 401,
		},
		{
			name: "iap",
			setup: func(t *testing.T, r *http.Request) {
				token := generateIAPToken(t, "https://example.com")
				r.Header.Add(iap.Header, token)
			},
			wantCode: 200,
		},
		{
			name: "org token",
			setup: func(t *testing.T, r *http.Request) {
				r.Header.Add("Authorization", "Bearer "+string(orgToken))
			},
			wantCode:    200,
			wantSubject: ot,
		},
		{
			name: "user token",
			setup: func(t *testing.T, r *http.Request) {
				_, token, err := daemon.Users.CreateToken(ctx, userpkg.CreateUserTokenOptions{
					Description: "lorem ipsum...",
				})
				require.NoError(t, err)
				r.Header.Add("Authorization", "Bearer "+string(token))
			},
			wantCode:    200,
			wantSubject: user,
		},
		{
			name: "site admin token",
			setup: func(t *testing.T, r *http.Request) {
				r.Header.Add("Authorization", "Bearer "+siteToken)
			},
			wantCode:    200,
			wantSubject: &userpkg.SiteAdmin,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/api/v2/protected", nil)
			if tt.setup != nil {
				tt.setup(t, r)
			}
			w := httptest.NewRecorder()
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				subj, err := authz.SubjectFromContext(r.Context())
				require.NoError(t, err)
				assert.Equal(t, tt.wantSubject, subj)
			})
			daemon.Tokens.Middleware.Authenticate(h).ServeHTTP(w, r)
			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

// generateIAPToken is a test helper to generate a token for a google cloud service
// account. If the necessary environment variable containing the service account
// key is not present then the test is skipped.
func generateIAPToken(t *testing.T, aud string) string {
	t.Helper()

	t.Log(os.Getenv("HTTPS_PROXY"))

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
	src, err := idtoken.NewTokenSource(
		t.Context(),
		aud,
		option.WithAuthCredentialsFile(option.ServiceAccount, path),
	)
	require.NoError(t, err)

	token, err := src.Token()
	require.NoError(t, err)
	return token.AccessToken
}
