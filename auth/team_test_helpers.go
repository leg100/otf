package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestTeam(t *testing.T, organization string) *otf.Team {
	return otf.NewTeam(otf.NewTeamOptions{uuid.NewString(), organization})
}

func CreateTestTeam(t *testing.T, db *pgdb, organization string) *otf.Team {
	ctx := context.Background()

	team := NewTestTeam(t, organization)
	err := db.createTeam(ctx, team)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.deleteTeam(ctx, team.ID)
	})
	return team
}

func newTestOwners(t *testing.T, organization string) *otf.Team {
	return otf.NewTeam(otf.NewTeamOptions{"owners", organization})
}

type fakeTeamApp struct {
	team    *otf.Team
	members []*otf.User

	teamService
}

func (f *fakeTeamApp) getTeamByID(ctx context.Context, teamID string) (*otf.Team, error) {
	return f.team, nil
}

func (f *fakeTeamApp) listTeams(ctx context.Context, organization string) ([]*otf.Team, error) {
	return []*otf.Team{f.team}, nil
}

func (f *fakeTeamApp) updateTeam(ctx context.Context, teamID string, opts otf.UpdateTeamOptions) (*otf.Team, error) {
	f.team.Update(opts)
	return f.team, nil
}

func (f *fakeTeamApp) listTeamMembers(ctx context.Context, teamID string) ([]*otf.User, error) {
	return f.members, nil
}
