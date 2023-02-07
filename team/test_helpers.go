package team

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestTeam(t *testing.T, organization string, opts ...NewTeamOption) *Team {
	return newTeam(uuid.NewString(), organization, opts...)
}

func NewTestOwners(t *testing.T, organization string, opts ...NewTeamOption) *Team {
	return newTeam("owners", organization, opts...)
}

func CreateTestTeam(t *testing.T, db otf.DB, organization string) *Team {
	ctx := context.Background()
	team := NewTestTeam(t, organization)
	teamDB := newDB(db)
	err := teamDB.CreateTeam(ctx, team)
	require.NoError(t, err)

	t.Cleanup(func() {
		teamDB.DeleteTeam(ctx, team.ID())
	})
	return team
}
