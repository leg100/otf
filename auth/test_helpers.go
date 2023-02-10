package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/stretchr/testify/require"
)

func NewTestUser(t *testing.T, opts ...NewUserOption) *User {
	return NewUser(uuid.NewString(), opts...)
}

func NewTestTeam(t *testing.T, organization string, opts ...NewTeamOption) *Team {
	return newTeam(uuid.NewString(), organization, opts...)
}

func NewTestOwners(t *testing.T, organization string, opts ...NewTeamOption) *Team {
	return newTeam("owners", organization, opts...)
}

func CreateTestUser(t *testing.T, db otf.DB, opts ...NewUserOption) *otf.User {
	ctx := context.Background()
	username := fmt.Sprintf("mr-%s", otf.GenerateRandomString(6))
	user := NewUser(username, opts...)
	userDB := newPGDB(db)

	err := userDB.CreateUser(ctx, user)
	require.NoError(t, err)

	t.Cleanup(func() {
		userDB.DeleteUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
	})
	return user
}

func CreateTestTeam(t *testing.T, db otf.DB, organization string) *Team {
	ctx := context.Background()
	team := NewTestTeam(t, organization)
	teamDB := newPGDB(db)
	err := teamDB.CreateTeam(ctx, team)
	require.NoError(t, err)

	t.Cleanup(func() {
		teamDB.DeleteTeam(ctx, team.ID())
	})
	return team
}

func createTestWorkspacePermission(t *testing.T, db DB, ws *Workspace, team Team, role rbac.Role) *otf.WorkspacePermission {
	ctx := context.Background()
	err := db.SetWorkspacePermission(ctx, ws.ID(), team.Name(), role)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.UnsetWorkspacePermission(ctx, ws.ID(), team.Name())
	})
	return &otf.WorkspacePermission{TeamID: team, Role: role}
}
package registry

import (
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestSession(t *testing.T, org *otf.Organization, opts ...NewTestSessionOption) *Session {
	session, err := newSession(org.Name())
	require.NoError(t, err)

	for _, o := range opts {
		o(session)
	}

	return session
}

type NewTestSessionOption func(*Session)

func OverrideTestRegistrySessionExpiry(expiry time.Time) NewTestSessionOption {
	return func(session *Session) {
		session.expiry = expiry
	}
}
package agenttoken

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestAgentToken(t *testing.T, org otf.Organization) *AgentToken {
	token, err := NewAgentToken(CreateAgentTokenOptions{
		Organization: org.Name(),
		Description:  "lorem ipsum...",
	})
	require.NoError(t, err)
	return token
}
package session

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createTestSession(t *testing.T, db otf.DB, userID string, opts ...otf.NewSessionOption) *otf.Session {
	session := NewTestSession(t, userID, opts...)
	ctx := context.Background()

	err := db.CreateSession(ctx, session)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteSession(ctx, session.Token())
	})
	return session
}

func NewTestSession(t *testing.T, userID string, opts ...NewSessionOption) *Session {
	session, err := NewSession(userID, "127.0.0.1")
	require.NoError(t, err)

	for _, o := range opts {
		o(session)
	}

	return session
}

type newTestDBOption func(*Options)

func overrideCleanupInterval(d time.Duration) newTestDBOption {
	return func(o *Options) {
		o.CleanupInterval = d
	}
}
