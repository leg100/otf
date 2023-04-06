package auth

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/leg100/otf/http/html/paths"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/require"
)

type fakeService struct {
	agentToken *AgentToken
	userToken  *Token
	team       *Team
	members    []*User

	token []byte

	AuthService
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

func (f *fakeService) GetTeamByID(ctx context.Context, teamID string) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) ListTeams(ctx context.Context, organization string) ([]*Team, error) {
	return []*Team{f.team}, nil
}

func (f *fakeService) UpdateTeam(ctx context.Context, teamID string, opts UpdateTeamOptions) (*Team, error) {
	f.team.Update(opts)
	return f.team, nil
}

func (f *fakeService) ListTeamMembers(ctx context.Context, teamID string) ([]*User, error) {
	return f.members, nil
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

func (f *fakeService) StartSession(w http.ResponseWriter, r *http.Request, opts StartUserSessionOptions) error {
	http.Redirect(w, r, paths.Profile(), http.StatusFound)
	return nil
}

func newTestJWT(t *testing.T, key string, kind authKind, lifetime time.Duration, claims ...string) string {
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
	serialized, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, []byte(key)))
	require.NoError(t, err)
	return string(serialized)
}

func newTestJWK(t *testing.T, secret string) jwk.Key {
	t.Helper()

	key, err := jwk.FromRaw([]byte(secret))
	require.NoError(t, err)
	return key
}
