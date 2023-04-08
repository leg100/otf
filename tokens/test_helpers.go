package tokens

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leg100/otf/http/html/paths"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/require"
)

type fakeService struct {
	agentToken *AgentToken
	userToken  *Token

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

func (f *fakeService) CreateToken(context.Context, CreateTokenOptions) (*Token, []byte, error) {
	return nil, f.token, nil
}

func (f *fakeService) ListTokens(context.Context) ([]*Token, error) {
	return []*Token{f.userToken}, nil
}

func (f *fakeService) DeleteToken(context.Context, string) error {
	return nil
}

func (f *fakeService) StartSession(w http.ResponseWriter, r *http.Request, opts StartSessionOptions) error {
	http.Redirect(w, r, paths.Profile(), http.StatusFound)
	return nil
}

func NewTestSessionJWT(t *testing.T, username, secret string, lifetime time.Duration) string {
	t.Helper()

	token, err := jwt.NewBuilder().
		Subject(username).
		IssuedAt(time.Now()).
		Claim("kind", userSessionKind).
		Expiration(time.Now().Add(lifetime)).
		Build()
	require.NoError(t, err)
	serialized, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, newTestJWK(t, secret)))
	require.NoError(t, err)
	return string(serialized)
}

func newTestJWT(t *testing.T, secret string, kind kind, lifetime time.Duration, claims ...string) string {
	t.Helper()

	builder := jwt.NewBuilder().
		IssuedAt(time.Now()).
		Claim("kind", kind).
		Expiration(time.Now().Add(lifetime))
	for i := 0; i < len(claims); i += 2 {
		builder = builder.Claim(claims[0], claims[1])
	}
	token, err := builder.Build()
	require.NoError(t, err)
	serialized, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, newTestJWK(t, secret)))
	require.NoError(t, err)
	return string(serialized)
}

func newTestJWK(t *testing.T, secret string) jwk.Key {
	t.Helper()

	key, err := jwk.FromRaw([]byte(secret))
	require.NoError(t, err)
	return key
}

func NewTestAgentToken(t *testing.T, org string) *AgentToken {
	token, _, err := NewAgentToken(NewAgentTokenOptions{
		CreateAgentTokenOptions: CreateAgentTokenOptions{
			Organization: org,
			Description:  "lorem ipsum...",
		},
		key: newTestJWK(t, "something_secret"),
	})
	require.NoError(t, err)
	return token
}

func NewTestToken(t *testing.T, org string) *Token {
	token, _, err := NewToken(NewTokenOptions{
		CreateTokenOptions: CreateTokenOptions{
			Description: "lorem ipsum...",
		},
		Username: uuid.NewString(),
		key:      newTestJWK(t, "something_secret"),
	})
	require.NoError(t, err)
	return token
}
