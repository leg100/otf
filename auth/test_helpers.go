package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leg100/otf"
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

type NewTestSessionOption func(*Session)

func OverrideTestRegistrySessionExpiry(expiry time.Time) NewTestSessionOption {
	return func(session *Session) {
		session.expiry = expiry
	}
}

func NewTestAgentToken(t *testing.T, org string) *AgentToken {
	token, err := newAgentToken(otf.CreateAgentTokenOptions{
		Organization: org,
		Description:  "lorem ipsum...",
	})
	require.NoError(t, err)
	return token
}
