package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func NewTestTeam(t *testing.T, organization string) *Team {
	return newTeam(createTeamOptions{uuid.NewString(), organization})
}

func newTestOwners(t *testing.T, organization string) *Team {
	return newTeam(createTeamOptions{"owners", organization})
}

func CreateTestTeam(t *testing.T, db *pgdb, organization string) *Team {
	ctx := context.Background()

	team := NewTestTeam(t, organization)
	err := db.createTeam(ctx, team)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.deleteTeam(ctx, team.ID())
	})
	return team
}
