package sql

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Create(t *testing.T) {
	db := newTestDB(t)
	user := newTestUser()

	defer db.UserStore().Delete(context.Background(), otf.UserSpec{Username: &user.Username})

	err := db.UserStore().Create(context.Background(), user)
	require.NoError(t, err)
}

func TestUser_Update_OrganizationMemberships(t *testing.T) {
	db := newTestDB(t)

	org1 := createTestOrganization(t, db)
	org2 := createTestOrganization(t, db)
	org3 := createTestOrganization(t, db)

	tests := []struct {
		name string
		// existing set of organization memberships
		existing []*otf.Organization
		// new set of organization memberships
		updated []*otf.Organization
	}{
		{
			name:     "from 0 to 3",
			existing: []*otf.Organization{},
			updated:  []*otf.Organization{org1, org2, org3},
		},
		{
			name:     "from 3 to 0",
			existing: []*otf.Organization{org1, org2, org3},
			updated:  nil,
		},
		{
			name:     "from 1 to 2",
			existing: []*otf.Organization{org1},
			updated:  []*otf.Organization{org2, org3},
		},
		{
			name:     "from 2 to 1",
			existing: []*otf.Organization{org1, org2},
			updated:  []*otf.Organization{org3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := createTestUser(t, db, withOrganizationMemberships(tt.existing...))

			user.Organizations = tt.updated

			err := db.UserStore().Update(context.Background(), otf.UserSpec{Username: &user.Username}, user)
			require.NoError(t, err)

			got, err := db.UserStore().Get(context.Background(), otf.UserSpec{Username: &user.Username})
			require.NoError(t, err)

			assert.Equal(t, tt.updated, got.Organizations)
		})
	}
}

func TestUser_Get(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	session := createTestSession(t, db, user.ID)
	token := createTestToken(t, db, user.ID, "testing")

	tests := []struct {
		name string
		spec otf.UserSpec
	}{
		{
			name: "username",
			spec: otf.UserSpec{Username: &user.Username},
		},
		{
			name: "session token",
			spec: otf.UserSpec{SessionToken: &session.Token},
		},
		{
			name: "auth token ID",
			spec: otf.UserSpec{AuthenticationTokenID: &token.ID},
		},
		{
			name: "auth token",
			spec: otf.UserSpec{AuthenticationToken: &token.Token},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.UserStore().Get(context.Background(), tt.spec)
			require.NoError(t, err)

			assert.Equal(t, got.ID, user.ID)
		})
	}

}

func TestUser_Get_WithSessions(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	_ = createTestSession(t, db, user.ID)
	_ = createTestSession(t, db, user.ID)

	got, err := db.UserStore().Get(context.Background(), otf.UserSpec{Username: &user.Username})
	require.NoError(t, err)

	assert.Equal(t, 2, len(got.Sessions))

}

// TestUser_SessionFlash demonstrates the session flash object is successfully
// serialized/deserialized from/to its struct
func TestUser_SessionFlash(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)

	t.Run("WithFlash", func(t *testing.T) {
		flash := &otf.Flash{
			Type:    otf.FlashSuccessType,
			Message: "test succeeded",
		}

		_ = createTestSession(t, db, user.ID, withFlash(flash))

		got, err := db.UserStore().Get(context.Background(), otf.UserSpec{Username: &user.Username})
		require.NoError(t, err)

		assert.Equal(t, flash, got.Sessions[0].Flash)
	})

	t.Run("WithNoFlash", func(t *testing.T) {
		_ = createTestSession(t, db, user.ID)

		got, err := db.UserStore().Get(context.Background(), otf.UserSpec{Username: &user.Username})
		require.NoError(t, err)

		assert.Nil(t, got.Sessions[0].Flash)
	})
}

func TestUser_List(t *testing.T) {
	db := newTestDB(t)
	user1 := createTestUser(t, db)
	user2 := createTestUser(t, db)
	user3 := createTestUser(t, db)

	users, err := db.UserStore().List(context.Background())
	require.NoError(t, err)

	assert.Contains(t, users, user1)
	assert.Contains(t, users, user2)
	assert.Contains(t, users, user3)
}

func TestUser_Delete(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)

	err := db.UserStore().Delete(context.Background(), otf.UserSpec{Username: &user.Username})
	require.NoError(t, err)

	// Verify zero users after deletion
	users, err := db.UserStore().List(context.Background())
	require.NoError(t, err)
	assert.NotContains(t, users, user)
}

func TestUser_CreateSession(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	session := newTestSession(t, user.ID)

	defer db.UserStore().DeleteSession(context.Background(), session.Token)

	err := db.UserStore().CreateSession(context.Background(), session)
	require.NoError(t, err)
}

func TestUser_UpdateSession(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	session := createTestSession(t, db, user.ID, withFlash(&otf.Flash{
		Type:    otf.FlashSuccessType,
		Message: "test succeeded",
	}))

	session.PopFlash()

	err := db.UserStore().UpdateSession(context.Background(), session.Token, session)
	require.NoError(t, err)

	// Verify session's flash has popped
	user, err = db.UserStore().Get(context.Background(), otf.UserSpec{SessionToken: &session.Token})
	require.NoError(t, err)
	assert.Nil(t, user.Sessions[0].Flash)
}

// TestUser_SessionCleanup tests the session cleanup background routine. We
// override the cleanup interval to just every 100ms, so after waiting for 300ms
// the sessions should be cleaned up.
func TestUser_SessionCleanup(t *testing.T) {
	db := newTestDB(t, 100*time.Millisecond)
	user := createTestUser(t, db)

	_ = createTestSession(t, db, user.ID, overrideExpiry(time.Now()))
	_ = createTestSession(t, db, user.ID, overrideExpiry(time.Now()))

	time.Sleep(300 * time.Millisecond)

	got, err := db.UserStore().Get(context.Background(), otf.UserSpec{Username: &user.Username})
	require.NoError(t, err)

	assert.Equal(t, 0, len(got.Sessions))
}

func TestUser_CreateToken(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	token, err := otf.NewToken(user.ID, "testing")
	require.NoError(t, err)

	defer db.UserStore().DeleteToken(context.Background(), token.ID)

	err = db.UserStore().CreateToken(context.Background(), token)
	require.NoError(t, err)
}

func TestUser_DeleteToken(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	token := createTestToken(t, db, user.ID, "testing")

	err := db.UserStore().DeleteToken(context.Background(), token.ID)
	require.NoError(t, err)
}

func TestDiffOrganizationLists(t *testing.T) {
	a := []*otf.Organization{
		{
			ID: "adidas",
		},
		{
			ID: "nike",
		},
	}
	b := []*otf.Organization{
		{
			ID: "adidas",
		},
		{
			ID: "puma",
		},
		{
			ID: "umbro",
		},
	}

	added, removed := diffOrganizationLists(a, b)

	assert.Equal(t, added, []*otf.Organization{
		{
			ID: "puma",
		},
		{
			ID: "umbro",
		},
	})
	assert.Equal(t, removed, []*otf.Organization{
		{
			ID: "nike",
		},
	})
}
