package sql

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrySession_Create(t *testing.T) {
	ctx := context.Background()

	db := newTestDB(t)
	org := createTestOrganization(t, db)

	session := otf.NewTestRegistrySession(t, org)

	err := db.CreateRegistrySession(ctx, session)
	require.NoError(t, err)
}

func TestRegistrySession_Get(t *testing.T) {
	ctx := context.Background()

	db := newTestDB(t)
	org := createTestOrganization(t, db)
	want := createTestRegistrySession(t, db, org)

	got, err := db.GetRegistrySession(ctx, want.Token())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestRegistrySession_Cleanup(t *testing.T) {
	ctx := context.Background()

	db := newTestDB(t, overrideCleanupInterval(100*time.Millisecond))
	org := createTestOrganization(t, db)

	session1 := createTestRegistrySession(t, db, org, otf.OverrideTestRegistrySessionExpiry(time.Now()))
	session2 := createTestRegistrySession(t, db, org, otf.OverrideTestRegistrySessionExpiry(time.Now()))

	time.Sleep(300 * time.Millisecond)

	_, err := db.GetRegistrySession(ctx, session1.Token())
	assert.Equal(t, otf.ErrResourceNotFound, err)

	_, err = db.GetRegistrySession(ctx, session2.Token())
	assert.Equal(t, otf.ErrResourceNotFound, err)
}
