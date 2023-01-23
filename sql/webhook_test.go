package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhook_Synchronise(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	repo := cloud.NewTestRepo()
	unsynced, err := otf.NewUnsynchronisedWebhook(otf.NewUnsynchronisedWebhookOptions{
		Identifier: repo.Identifier,
		Cloud:      "github",
	})
	require.NoError(t, err)

	// first sync creates hook in the DB and sets cloud ID
	got, err := db.SynchroniseWebhook(ctx, unsynced, func(hook *otf.Webhook) (string, error) {
		require.Nil(t, hook) // there should be no existing hook in DB
		return "123", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "123", got.VCSID())
	assert.Equal(t, unsynced, got.UnsynchronisedWebhook)

	// second sync retrieves existing hook in the DB and updates its cloud ID
	// (to mimic a hook having been re-created on the cloud and hence a new ID
	// has been generated)
	updated, err := db.SynchroniseWebhook(ctx, unsynced, func(hook *otf.Webhook) (string, error) {
		assert.Equal(t, got, hook) // should be the same hook created first time round
		return "456", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "456", updated.VCSID())                  // updated cloud ID
	assert.Equal(t, unsynced, updated.UnsynchronisedWebhook) // rest of hook is the same
}

func TestWebhook_Get(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	repo := cloud.NewTestRepo()
	cc := github.Defaults()

	want := createTestWebhook(t, db, repo, cc)

	got, err := db.GetWebhook(ctx, want.ID())
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestWebhook_Delete(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	repo := cloud.NewTestRepo()
	cc := github.Defaults()

	hook1 := createTestWebhook(t, db, repo, cc)
	// second call to create shouldn't create a hook but instead increments the
	// 'connected' field and returns the same hook
	hook2 := createTestWebhook(t, db, repo, cc)
	assert.Equal(t, hook1, hook2)

	// first call to delete should decrement connected field and return an error
	// indicating hook is still connected.
	_, err := db.DeleteWebhook(ctx, hook1.ID())
	require.Equal(t, otf.ErrWebhookConnected, err)

	// second call to delete should decrement connected field down to zero and
	// now the hook is deleted.
	_, err = db.DeleteWebhook(ctx, hook2.ID())
	assert.NoError(t, err)
}
