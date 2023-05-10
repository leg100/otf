package tokens

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/testutils"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/require"
)

type fakeService struct {
	agentToken *AgentToken
	userToken  *UserToken

	token []byte

	TokensService
}

func (f *fakeService) CreateAgentToken(context.Context, CreateAgentTokenOptions) ([]byte, error) {
	return f.token, nil
}

func (f *fakeService) ListAgentTokens(context.Context, string) ([]*AgentToken, error) {
	return []*AgentToken{f.agentToken}, nil
}

func (f *fakeService) DeleteAgentToken(context.Context, string) (*AgentToken, error) {
	return f.agentToken, nil
}

func (f *fakeService) CreateUserToken(context.Context, CreateUserTokenOptions) (*UserToken, []byte, error) {
	return nil, f.token, nil
}

func (f *fakeService) ListUserTokens(context.Context) ([]*UserToken, error) {
	return []*UserToken{f.userToken}, nil
}

func (f *fakeService) DeleteUserToken(context.Context, string) error {
	return nil
}

func (f *fakeService) StartSession(w http.ResponseWriter, r *http.Request, opts StartSessionOptions) error {
	http.Redirect(w, r, paths.Profile(), http.StatusFound)
	return nil
}

func NewTestSessionJWT(t *testing.T, username string, secret []byte, lifetime time.Duration) string {
	t.Helper()

	return NewTestJWT(t, secret, userSessionKind, lifetime, "sub", username)
}

func NewTestJWT(t *testing.T, secret []byte, kind kind, lifetime time.Duration, claims ...string) string {
	t.Helper()

	claimsMap := make(map[string]string, len(claims)/2)
	for i := 0; i < len(claims); i += 2 {
		claimsMap[claims[i]] = claims[i+1]
	}
	token, err := newToken(newTokenOptions{
		key:    newTestJWK(t, secret),
		kind:   kind,
		expiry: internal.Time(time.Now().Add(lifetime)),
		claims: claimsMap,
	})
	require.NoError(t, err)
	return string(token)
}

func newTestJWK(t *testing.T, secret []byte) jwk.Key {
	t.Helper()

	key, err := jwk.FromRaw(secret)
	require.NoError(t, err)
	return key
}

func NewTestAgentToken(t *testing.T, org string) *AgentToken {
	token, _, err := NewAgentToken(NewAgentTokenOptions{
		CreateAgentTokenOptions: CreateAgentTokenOptions{
			Organization: org,
			Description:  "lorem ipsum...",
		},
		key: newTestJWK(t, testutils.NewSecret(t)),
	})
	require.NoError(t, err)
	return token
}

func NewTestToken(t *testing.T, org string) *UserToken {
	token, _, err := NewUserToken(NewUserTokenOptions{
		CreateUserTokenOptions: CreateUserTokenOptions{
			Description: "lorem ipsum...",
		},
		Username: uuid.NewString(),
		key:      newTestJWK(t, testutils.NewSecret(t)),
	})
	require.NoError(t, err)
	return token
}
