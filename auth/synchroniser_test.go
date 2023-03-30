package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSyncTeams(t *testing.T) {
	ctx := context.Background()
	svc := &fakeSynchroniserService{}

	team1 := NewTestTeam(t, "org-1")
	team2 := NewTestTeam(t, "org-1")
	team3 := NewTestTeam(t, "org-1")

	user := NewUser(uuid.NewString(), WithTeams(team1, team2))

	s := &synchroniser{
		AuthService: svc,
	}
	want := []*Team{team2, team3}
	err := s.syncTeams(ctx, user, want)
	require.NoError(t, err)

	// expect membership to have been added to team3
	if assert.Equal(t, 1, len(svc.addedTeams)) {
		assert.Equal(t, team3.ID, svc.addedTeams[0])
	}
	// expect membership to have been removed from team1
	if assert.Equal(t, 1, len(svc.removedTeams)) {
		assert.Equal(t, team1.ID, svc.removedTeams[0])
	}
}

type fakeSynchroniserService struct {
	addedTeams, removedTeams []string

	AuthService
}

func (f *fakeSynchroniserService) AddTeamMembership(ctx context.Context, _, team string) error {
	f.addedTeams = append(f.addedTeams, team)
	return nil
}

func (f *fakeSynchroniserService) RemoveTeamMembership(ctx context.Context, _, team string) error {
	f.removedTeams = append(f.removedTeams, team)
	return nil
}
