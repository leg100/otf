package sql

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/github"
	"github.com/stretchr/testify/require"
)

func TestWebhook_CreateUnsynchronised(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	repo := cloud.NewTestRepo()
	want, err := otf.NewUnsynchronisedWebhook(otf.NewUnsynchronisedWebhookOptions{
		Identifier:  repo.Identifier,
		CloudConfig: github.Defaults(),
	})
	require.NoError(t, err)

	got, err := db.CreateUnsynchronisedWebhook(ctx, want)
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestWebhook_SynchroniseWebhook(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	repo := cloud.NewTestRepo()
	cc := github.Defaults()
	unsynced := createTestUnsynchronisedWebhook(t, db, repo, cc)

	cloudID := uuid.NewString()
	hook, err := db.SynchroniseWebhook(ctx, unsynced.ID(), cloudID)
	require.NoError(t, err)
	require.Equal(t, hook.VCSID(), cloudID)
}

func TestWebhook_DeleteWebhook(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	repo := cloud.NewTestRepo()
	cc := github.Defaults()
	hook1 := createTestWebhook(t, db, repo, cc)
	_ = createTestWebhook(t, db, repo, cc)

	err := db.DeleteWebhook(ctx, hook1.ID())
	require.NoError(t, err)
}
