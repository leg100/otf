package auth

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/require"
)

func NewTestTeam(t *testing.T, organization string) *Team {
	return NewTeam(NewTeamOptions{uuid.NewString(), organization})
}

func CreateTestTeam(t *testing.T, db otf.DB, organization *organization.Organization) *Team {
	userDB := newDB(db, logr.Discard())
	return createTestTeam(t, userDB, organization.Name)
}

func createTestTeam(t *testing.T, db *pgdb, organization string) *Team {
	ctx := context.Background()

	team := NewTestTeam(t, organization)
	err := db.createTeam(ctx, team)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.deleteTeam(ctx, team.ID)
	})
	return team
}

func newTestOwners(t *testing.T, organization string) *Team {
	return NewTeam(NewTeamOptions{"owners", organization})
}

type fakeTeamApp struct {
	team    *Team
	members []*User

	teamService
}

func (f *fakeTeamApp) getTeamByID(ctx context.Context, teamID string) (*Team, error) {
	return f.team, nil
}

func (f *fakeTeamApp) listTeams(ctx context.Context, organization string) ([]*Team, error) {
	return []*Team{f.team}, nil
}

func (f *fakeTeamApp) updateTeam(ctx context.Context, teamID string, opts UpdateTeamOptions) (*Team, error) {
	f.team.Update(opts)
	return f.team, nil
}

func (f *fakeTeamApp) listTeamMembers(ctx context.Context, teamID string) ([]*User, error) {
	return f.members, nil
}
