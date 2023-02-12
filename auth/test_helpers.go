package auth

import (
	"context"
	"net/http/httptest"
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

func CreateTestTeam(t *testing.T, db db, organization string) *Team {
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

func NewTestAgentToken(t *testing.T, org string) *agentToken {
	token, err := newAgentToken(otf.CreateAgentTokenOptions{
		Organization: org,
		Description:  "lorem ipsum...",
	})
	require.NoError(t, err)
	return token
}

func createTestSession(t *testing.T, db db, userID string, expiry *time.Time) *Session {
	ctx := context.Background()

	session := newTestSession(t, userID, expiry)
	err := db.createSession(ctx, session)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.deleteSession(ctx, session.Token())
	})
	return session
}

func newTestSession(t *testing.T, userID string, expiry *time.Time) *Session {
	r := httptest.NewRequest("", "/", nil)
	session, err := newSession(r, userID)
	require.NoError(t, err)
	if expiry != nil {
		session.expiry = *expiry
	}

	return session
}
