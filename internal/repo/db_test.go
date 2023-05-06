package repo

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	t.Run("create hook", func(t *testing.T) {
		want := newTestHook(t, db.factory, otf.String("123"))

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
		hook := newTestHook(t, db.factory, otf.String("123"))
		want, err := db.getOrCreateHook(ctx, hook)
		require.NoError(t, err)

		got, err := db.getHookByID(ctx, want.id)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("delete hook", func(t *testing.T) {
		hook := newTestHook(t, db.factory, otf.String("123"))
		_, err := db.getOrCreateHook(ctx, hook)
		require.NoError(t, err)

		_, err = db.deleteHook(ctx, hook.id)
		require.NoError(t, err)
	})
}
