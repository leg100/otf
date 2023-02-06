package registry

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrySession_Create(t *testing.T) {
	ctx := context.Background()

	db := sql.NewTestDB(t)
	sessionDB := newDB(ctx, db, 0)
	org := sql.CreateTestOrganization(t, db)

	session := NewTestSession(t, org)

	err := sessionDB.create(ctx, session)
	require.NoError(t, err)
}

func TestRegistrySession_Get(t *testing.T) {
	ctx := context.Background()

	db := sql.NewTestDB(t)
	sessionDB := newDB(ctx, db, 0)
	org := sql.CreateTestOrganization(t, db)
	want := createTestSession(t, sessionDB, org)

	got, err := sessionDB.get(ctx, want.Token())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestRegistrySession_Cleanup(t *testing.T) {
	ctx := context.Background()

	db := sql.NewTestDB(t)
	sessionDB := newDB(ctx, db, 100*time.Millisecond)
	org := sql.CreateTestOrganization(t, db)

	session1 := createTestSession(t, sessionDB, org, OverrideTestRegistrySessionExpiry(time.Now()))
	session2 := createTestSession(t, sessionDB, org, OverrideTestRegistrySessionExpiry(time.Now()))

	time.Sleep(300 * time.Millisecond)

	_, err := sessionDB.get(ctx, session1.Token())
	assert.Equal(t, otf.ErrResourceNotFound, err)

	_, err = sessionDB.get(ctx, session2.Token())
	assert.Equal(t, otf.ErrResourceNotFound, err)
}

func createTestSession(t *testing.T, sessionDB db, org *otf.Organization, opts ...NewTestSessionOption) *Session {
	ctx := context.Background()

	session := NewTestSession(t, org, opts...)

	err := sessionDB.create(ctx, session)
	require.NoError(t, err)

	return session
}
