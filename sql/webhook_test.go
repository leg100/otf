package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhook_Sync_Create(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	repo := otf.NewTestRepo()
	want := otf.NewTestWebhook(repo)

	createFunc := func(context.Context, otf.WebhookCreatorOptions) (*otf.Webhook, error) {
		return want, nil
	}
	opts := otf.SyncWebhookOptions{
		CreateWebhookFunc: createFunc,
		HTTPURL:           want.HTTPURL,
	}

	got, err := db.SyncWebhook(ctx, opts)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestWebhook_Sync_NoChange(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	want := createTestWebhook(t, db)

	updateFunc := func(context.Context, otf.WebhookUpdaterOptions) (string, error) {
		return want.VCSID, nil
	}
	opts := otf.SyncWebhookOptions{
		UpdateWebhookFunc: updateFunc,
		HTTPURL:           want.HTTPURL,
	}

	got, err := db.SyncWebhook(ctx, opts)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestWebhook_GetSecret(t *testing.T) {
	db := newTestDB(t)
	hook := createTestWebhook(t, db)

	got, err := db.GetWebhookSecret(context.Background(), hook.WebhookID)
	require.NoError(t, err)
	assert.Equal(t, "secret", got)
}

func TestWebhook_Delete(t *testing.T) {
	db := newTestDB(t)
	hook := createTestWebhook(t, db)

	err := db.DeleteWebhook(context.Background(), hook.WebhookID)
	require.NoError(t, err)
}
