package authenticator

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSyncTeams(t *testing.T) {
	ctx := context.Background()
	svc := &fakeAuthService{}

	team1 := auth.NewTestTeam(t, "org-1")
	team2 := auth.NewTestTeam(t, "org-1")
	team3 := auth.NewTestTeam(t, "org-1")

	user := auth.NewUser(uuid.NewString(), auth.WithTeams(team1, team2))

	s := &synchroniser{
		AuthService: svc,
	}
	want := []*auth.Team{team2, team3}
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
