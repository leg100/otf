package auth

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrySession(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	t.Run("create", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		session, err := NewRegistrySession(org.Name)
		require.NoError(t, err)

		err = db.createRegistrySession(ctx, session)
		require.NoError(t, err)
	})

	t.Run("get", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		want := createTestRegistrySession(t, db, org.Name, nil)

		got, err := db.getRegistrySession(ctx, want.Token)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("cleanup", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)

		session1 := createTestRegistrySession(t, db, org.Name, otf.Time(time.Now()))
		session2 := createTestRegistrySession(t, db, org.Name, otf.Time(time.Now()))

		_, err := db.DeleteExpiredRegistrySessions(ctx)
		require.NoError(t, err)

		_, err = db.getRegistrySession(ctx, session1.Token)
		assert.Equal(t, otf.ErrResourceNotFound, err)

		_, err = db.getRegistrySession(ctx, session2.Token)
		assert.Equal(t, otf.ErrResourceNotFound, err)
	})
}
