package user

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
	return &otf.WorkspacePermission{Team: team, Role: role}
}
