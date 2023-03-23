package auth

import (
	"context"
	"net/http"
	"time"
)

type NewTestSessionOption func(*Session)

func OverrideTestRegistrySessionExpiry(expiry time.Time) NewTestSessionOption {
	return func(session *Session) {
		session.expiry = expiry
	}
}

type fakeService struct {
	sessions   []*Session
	agentToken *AgentToken
	team       *Team
	members    []*User
	token      *Token

	AuthService
}

func (f *fakeService) listSessions(context.Context, string) ([]*Session, error) {
	return f.sessions, nil
}

func (f *fakeService) deleteSession(context.Context, string) error {
	return nil
}

func (f *fakeService) createSession(*http.Request, string) (*Session, error) {
	return &Session{}, nil
}

func (f *fakeService) CreateAgentToken(context.Context, CreateAgentTokenOptions) (*AgentToken, error) {
	return f.agentToken, nil
}

func (f *fakeService) listAgentTokens(context.Context, string) ([]*AgentToken, error) {
	return []*AgentToken{f.agentToken}, nil
}

func (f *fakeService) deleteAgentToken(context.Context, string) (*AgentToken, error) {
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

func (f *fakeService) CreateToken(context.Context, string, *TokenCreateOptions) (*Token, error) {
	return f.token, nil
}

func (f *fakeService) ListTokens(context.Context, string) ([]*Token, error) {
	return []*Token{f.token}, nil
}

func (f *fakeService) DeleteToken(context.Context, string, string) error {
	return nil
}
