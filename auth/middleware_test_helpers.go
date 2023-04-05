package auth

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeMiddlewareService struct {
	agentToken    string
	registryToken string
	sessionToken  string
	userToken     string
}

func (f *fakeMiddlewareService) GetAgentToken(ctx context.Context, token string) (*AgentToken, error) {
	if f.agentToken == token {
		return nil, nil
	}
	return nil, errors.New("invalid")
}

func (f *fakeMiddlewareService) GetRegistrySession(ctx context.Context, token string) (*RegistrySession, error) {
	if f.registryToken == token {
		return nil, nil
	}
	return nil, errors.New("invalid")
}

func (f *fakeMiddlewareService) GetUser(ctx context.Context, spec UserSpec) (*User, error) {
	if spec.AuthenticationToken != nil {
		if f.userToken == *spec.AuthenticationToken {
			return nil, nil
		}
	} else if spec.Username != nil {
		// this is the google jwt check, so accept any username
		return nil, nil
	}
	return nil, errors.New("invalid")
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
		key, ok := os.LookupEnv("GOOGLE_CREDENTIALS")
		if !ok {
			t.Skip("Export a valid GOOGLE_APPLICATION_CREDENTIALS or GOOGLE_CREDENTIALS before running this test")
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
