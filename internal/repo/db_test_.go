package repo

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDB should not be called directly but via the 'integration' package (Hence
// why this file is suffixed with '_' to prevent Go from detecting it as a test
// file).
func TestDB(t *testing.T, sqldb *sql.DB) {
	ctx := context.Background()
	db := newTestDB(t, sqldb)

	t.Run("create hook", func(t *testing.T) {
		want := newTestHook(t, db.factory, internal.String("123"))

		got, err := db.getOrCreateHook(ctx, want)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("update hook cloud ID", func(t *testing.T) {
		hook := newTestHook(t, db.factory, nil)
		want, err := db.getOrCreateHook(ctx, hook)
		require.NoError(t, err)

		err = db.updateHookCloudID(ctx, hook.id, "123")
		require.NoError(t, err)

		got, err := db.getHookByID(ctx, want.id)
		require.NoError(t, err)
		assert.Equal(t, "123", *got.cloudID)
	})

	t.Run("get hook", func(t *testing.T) {
		hook := newTestHook(t, db.factory, internal.String("123"))
		want, err := db.getOrCreateHook(ctx, hook)
		require.NoError(t, err)

		got, err := db.getHookByID(ctx, want.id)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("delete hook", func(t *testing.T) {
		hook := newTestHook(t, db.factory, internal.String("123"))
		_, err := db.getOrCreateHook(ctx, hook)
		require.NoError(t, err)

		_, err = db.deleteHook(ctx, hook.id)
		require.NoError(t, err)
	})
}
