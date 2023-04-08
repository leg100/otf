package tokens

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/idtoken"
)

type fakeMiddlewareService struct {
	auth.AuthService
	TokensService
}

func (f *fakeMiddlewareService) GetAgentToken(ctx context.Context, token string) (*AgentToken, error) {
	return &AgentToken{}, nil
}

func (f *fakeMiddlewareService) GetUser(ctx context.Context, spec auth.UserSpec) (*auth.User, error) {
	return &auth.User{}, nil
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

func fakeTokenMiddleware(t *testing.T, secret string) mux.MiddlewareFunc {
	t.Helper()

	key := newTestJWK(t, secret)
	return newMiddleware(middlewareOptions{
		AuthService:       &fakeMiddlewareService{},
		AgentTokenService: &fakeMiddlewareService{},
		key:               key,
	})
}

func fakeSiteTokenMiddleware(t *testing.T, token string) mux.MiddlewareFunc {
	t.Helper()

	key := newTestJWK(t, "abcdef123") // not used but constructor requires it
	return newMiddleware(middlewareOptions{
		AuthService:       &fakeMiddlewareService{},
		AgentTokenService: &fakeMiddlewareService{},
		SiteToken:         token,
		key:               key,
	})
}

func fakeIAPMiddleware(t *testing.T, aud string) mux.MiddlewareFunc {
	t.Helper()

	key := newTestJWK(t, "abcdef123") // not used but constructor requires it
	return newMiddleware(middlewareOptions{
		AuthService:       &fakeMiddlewareService{},
		AgentTokenService: &fakeMiddlewareService{},
		GoogleIAPConfig: GoogleIAPConfig{
			Audience: aud,
		},
		key: key,
	})
}

var emptyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// implicitly responds with 200 OK
})

func wantSubjectHandler(t *testing.T, want any) http.HandlerFunc {
	t.Helper()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, err := otf.SubjectFromContext(r.Context())
		require.NoError(t, err)
		if assert.NotNil(t, got, "subject is missing") {
			assert.Equal(t, reflect.TypeOf(want), reflect.TypeOf(got))
		}
	})
}

func newIAPToken(t *testing.T, aud string) string {
	t.Helper()

	// tests are sometimes run behind an http proxy with a self-signed cert,
	// which the google oauth2 client fails to verify, so just for this test do
	// not use the proxy.
	t.Setenv("HTTPS_PROXY", "")

	ctx := context.Background()
	credspath := getGoogleCredentialsPath(t)
	src, err := idtoken.NewTokenSource(ctx, aud, idtoken.WithCredentialsFile(credspath))
	require.NoError(t, err)

	token, err := src.Token()
	require.NoError(t, err)
	return token.AccessToken
}
