package auth

import (
	"context"
	"testing"

	"github.com/leg100/otf/http/html"
	"github.com/stretchr/testify/require"
)

type fakeService struct {
	team    *Team
	members []*User

	AuthService
}

func (f *fakeService) GetTeamByID(ctx context.Context, teamID string) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) ListUsers(ctx context.Context) ([]*User, error) {
	return nil, nil
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

func (f *fakeService) DeleteTeam(ctx context.Context, teamID string) error {
	return nil
}

func newFakeWeb(t *testing.T, svc AuthService) *webHandlers {
	t.Helper()

	renderer, err := html.NewRenderer(false)
	require.NoError(t, err)
	return &webHandlers{
		svc:      svc,
		Renderer: renderer,
	}
}
