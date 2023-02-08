package user

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/team"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Create(t *testing.T) {
	db := sql.NewTestDB(t)
	userDB := newPGDB(db)

	user := NewUser("mr-t")

	defer userDB.DeleteUser(context.Background(), UserSpec{Username: otf.String(user.Username())})

	err := userDB.CreateUser(context.Background(), user)
	require.NoError(t, err)
}

func TestUser_AddOrganizationMembership(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	userDB := newPGDB(db)

	org := organization.CreateTestOrganization(t, db)
	user := CreateTestUser(t, db)

	err := userDB.AddOrganizationMembership(ctx, user.ID(), org.Name())
	require.NoError(t, err)

	got, err := userDB.GetUser(ctx, UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.Contains(t, got.Organizations(), org.Name())
}

func TestUser_RemoveOrganizationMembership(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	userDB := newPGDB(db)

	org := organization.CreateTestOrganization(t, db)
	user := CreateTestUser(t, db, WithOrganizationMemberships(org.Name()))

	err := userDB.RemoveOrganizationMembership(ctx, user.ID(), org.Name())
	require.NoError(t, err)

	got, err := userDB.GetUser(ctx, UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.NotContains(t, got.Organizations(), org)
}

func TestUser_AddTeamMembership(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	userDB := newPGDB(db)

	org := organization.CreateTestOrganization(t, db)
	team := team.CreateTestTeam(t, db, org.Name())
	user := CreateTestUser(t, db, WithOrganizationMemberships(org.Name()))

	err := userDB.AddTeamMembership(ctx, user.ID(), team.ID())
	require.NoError(t, err)

	got, err := userDB.GetUser(ctx, UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.Contains(t, got.Teams(), team)
}

func TestUser_RemoveTeamMembership(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	userDB := newPGDB(db)

	org := organization.CreateTestOrganization(t, db)
	team := team.CreateTestTeam(t, db, org.Name())
	user := CreateTestUser(t, db, WithOrganizationMemberships(org.Name()), WithTeamMemberships(team))

	err := userDB.RemoveTeamMembership(ctx, user.ID(), team.ID())
	require.NoError(t, err)

	got, err := userDB.GetUser(ctx, UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.NotContains(t, got.Teams(), team)
}

func TestTeam_ListTeamMembers(t *testing.T) {
	db := sql.NewTestDB(t)
	userDB := newPGDB(db)

	org := organization.CreateTestOrganization(t, db)
	team := team.CreateTestTeam(t, db, org.Name())

	memberships := []NewUserOption{
		WithOrganizationMemberships(org.Name()),
		WithTeamMemberships(team),
	}
	user1 := CreateTestUser(t, db, memberships...)
	user2 := CreateTestUser(t, db, memberships...)

	got, err := userDB.ListTeamMembers(context.Background(), team.ID())
	require.NoError(t, err)

	assert.Contains(t, got, user1)
	assert.Contains(t, got, user2)
}

func TestUser_Get(t *testing.T) {
	db := sql.NewTestDB(t)
	userDB := newPGDB(db)

	org1 := organization.CreateTestOrganization(t, db)
	org2 := organization.CreateTestOrganization(t, db)
	team1 := team.CreateTestTeam(t, db, org1.Name())
	team2 := team.CreateTestTeam(t, db, org2.Name())

	user := CreateTestUser(t, db,
		WithOrganizationMemberships(org1.Name(), org2.Name()),
		WithTeamMemberships(team1, team2))

	session1 := createTestSession(t, db, user.ID())
	_ = createTestSession(t, db, user.ID())

	token1 := createTestToken(t, db, user.ID(), "testing")
	_ = createTestToken(t, db, user.ID(), "testing")

	tests := []struct {
		name string
		spec UserSpec
	}{
		{
			name: "id",
			spec: UserSpec{UserID: otf.String(user.ID())},
		},
		{
			name: "username",
			spec: UserSpec{Username: otf.String(user.Username())},
		},
		{
			name: "session token",
			spec: UserSpec{SessionToken: otf.String(session1.Token())},
		},
		{
			name: "auth token",
			spec: UserSpec{AuthenticationToken: otf.String(token1.Token())},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userDB.GetUser(context.Background(), tt.spec)
			require.NoError(t, err)

			assert.Equal(t, got.ID(), user.ID())
			assert.Equal(t, got.Username(), user.Username())
			assert.Equal(t, got.CreatedAt(), user.CreatedAt())
			assert.Equal(t, got.UpdatedAt(), user.UpdatedAt())
			assert.Equal(t, 2, len(got.Organizations()))
			assert.Equal(t, 2, len(got.Teams()))
		})
	}
}

func TestUser_Get_NotFound(t *testing.T) {
	db := sql.NewTestDB(t)
	userDB := newPGDB(db)

	_, err := userDB.GetUser(context.Background(), UserSpec{Username: otf.String("does-not-exist")})
	assert.Equal(t, otf.ErrResourceNotFound, err)
}

func TestUser_List(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	userDB := newPGDB(db)

	org := organization.CreateTestOrganization(t, db)
	team := team.CreateTestTeam(t, db, org.Name())
	user1 := CreateTestUser(t, db)
	user2 := CreateTestUser(t, db, WithOrganizationMemberships(org.Name()))
	user3 := CreateTestUser(t, db, WithOrganizationMemberships(org.Name()), WithTeamMemberships(team))

	// Retrieve all users
	users, err := userDB.ListUsers(ctx, UserListOptions{})
	require.NoError(t, err)

	require.Contains(t, users, user1)
	assert.Contains(t, users, user2)
	assert.Contains(t, users, user3)

	// Retrieve users in org
	users, err = userDB.ListUsers(ctx, UserListOptions{Organization: otf.String(org.Name())})
	require.NoError(t, err)

	assert.NotContains(t, users, user1)
	assert.Contains(t, users, user2)
	assert.Contains(t, users, user3)

	// Retrieve users in org belonging to team
	users, err = userDB.ListUsers(ctx, UserListOptions{
		Organization: otf.String(org.Name()),
		TeamName:     otf.String(team.Name()),
	})
	require.NoError(t, err)

	assert.NotContains(t, users, user1)
	assert.NotContains(t, users, user2)
	assert.Contains(t, users, user3)
}

func TestUser_Delete(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	userDB := newPGDB(db)

	user := CreateTestUser(t, db)

	err := userDB.DeleteUser(ctx, UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	users, err := userDB.ListUsers(ctx, UserListOptions{})
	require.NoError(t, err)
	assert.NotContains(t, users, user)
}

func TestTeam_Create(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	teamDB := newPGDB(db)
	org := organization.CreateTestOrganization(t, db)
	team := newTeam("team-awesome", org.Name())

	defer teamDB.DeleteTeam(ctx, team.ID())

	err := teamDB.CreateTeam(ctx, team)
	require.NoError(t, err)
}

func TestTeam_Update_ByID(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	teamDB := newPGDB(db)

	org := organization.CreateTestOrganization(t, db)
	team := CreateTestTeam(t, teamDB, org.Name())

	_, err := teamDB.UpdateTeam(ctx, team.ID(), func(team *Team) error {
		return team.Update(UpdateTeamOptions{
			OrganizationAccess: OrganizationAccess{
				ManageWorkspaces: true,
				ManageVCS:        true,
				ManageRegistry:   true,
			},
		})
	})
	require.NoError(t, err)

	got, err := teamDB.GetTeam(ctx, team.Name(), org.Name())
	require.NoError(t, err)

	assert.True(t, got.OrganizationAccess().ManageWorkspaces)
	assert.True(t, got.OrganizationAccess().ManageVCS)
	assert.True(t, got.OrganizationAccess().ManageRegistry)
}

func TestTeam_Get(t *testing.T) {
	db := sql.NewTestDB(t)
	teamDB := newPGDB(db)

	org := organization.CreateTestOrganization(t, db)
	team := CreateTestTeam(t, teamDB, org.Name())

	got, err := teamDB.GetTeam(context.Background(), team.Name(), org.Name())
	require.NoError(t, err)

	assert.Equal(t, team, got)
}

func TestTeam_GetByID(t *testing.T) {
	db := sql.NewTestDB(t)
	teamDB := newPGDB(db)

	org := organization.CreateTestOrganization(t, db)
	want := CreateTestTeam(t, db, org.Name())

	got, err := teamDB.GetTeamByID(context.Background(), want.ID())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestTeam_List(t *testing.T) {
	db := sql.NewTestDB(t)
	org := organization.CreateTestOrganization(t, db)
	teamDB := newPGDB(db)

	team1 := CreateTestTeam(t, db, org.Name())
	team2 := CreateTestTeam(t, db, org.Name())
	team3 := CreateTestTeam(t, db, org.Name())

	got, err := teamDB.ListTeams(context.Background(), org.Name())
	require.NoError(t, err)

	assert.Contains(t, got, team1)
	assert.Contains(t, got, team2)
	assert.Contains(t, got, team3)
}

func TestWorkspacePermissions_Set(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	team := createTestTeam(t, db, org)

	err := db.SetWorkspacePermission(ctx, ws.ID(), team.Name(), rbac.WorkspacePlanRole)
	require.NoError(t, err)

	t.Run("Update", func(t *testing.T) {
		err := db.SetWorkspacePermission(ctx, ws.ID(), team.Name(), rbac.WorkspaceAdminRole)
		require.NoError(t, err)
	})
}

func TestWorkspacePermissions_List(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	team1 := createTestTeam(t, db, org)
	team2 := createTestTeam(t, db, org)
	perm1 := createTestWorkspacePermission(t, db, ws, team1, rbac.WorkspaceAdminRole)
	perm2 := createTestWorkspacePermission(t, db, ws, team2, rbac.WorkspacePlanRole)

	perms, err := db.ListWorkspacePermissions(ctx, ws.ID())
	require.NoError(t, err)
	if assert.Equal(t, 2, len(perms)) {
		assert.Contains(t, perms, perm1)
		assert.Contains(t, perms, perm2)
	}
}

func TestWorkspacePermissions_Unset(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	team := createTestTeam(t, db, org)
	_ = createTestWorkspacePermission(t, db, ws, team, rbac.WorkspaceAdminRole)

	err := db.UnsetWorkspacePermission(ctx, ws.ID(), team.Name())
	require.NoError(t, err)

	perms, err := db.ListWorkspacePermissions(ctx, ws.ID())
	require.NoError(t, err)
	assert.Equal(t, 0, len(perms))
}
