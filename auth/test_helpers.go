package auth

import (
	"context"
	"net/http"

	"github.com/leg100/otf/http/html/paths"
)

type fakeService struct {
	agentToken *AgentToken
	team       *Team
	members    []*User
	token      *Token

	AuthService
}

func (f *fakeService) CreateAgentToken(context.Context, CreateAgentTokenOptions) ([]byte, error) {
	return f.agentToken, nil
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

func (f *fakeService) CreateToken(context.Context, *TokenCreateOptions) (*Token, error) {
	return f.token, nil
}

func (f *fakeService) ListTokens(context.Context) ([]*Token, error) {
	return []*Token{f.token}, nil
}

func (f *fakeService) DeleteToken(context.Context, string) error {
	return nil
}

func (f *fakeService) StartSession(w http.ResponseWriter, r *http.Request, opts StartUserSessionOptions) error {
	http.Redirect(w, r, paths.Profile(), http.StatusFound)
	return nil
}
