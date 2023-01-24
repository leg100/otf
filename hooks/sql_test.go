package hooks

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	want := newTestHook(t, db.factory, otf.String("123"))

	got, err := db.create(ctx, want, &fakeCloudClient{hook: cloud.Webhook{ID: "123"}})
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	unsynced := newTestHook(t, db.factory, nil)

	want, err := db.create(ctx, unsynced, &fakeCloudClient{hook: cloud.Webhook{ID: "123"}})
	require.NoError(t, err)

	got, err := db.get(ctx, want.id)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	client := &fakeCloudClient{hook: cloud.Webhook{ID: "123"}}
	hook := newTestHook(t, db.factory, nil)

	// first create sets connected=1
	_, err := db.create(ctx, hook, client)
	require.NoError(t, err)

	// second create sets connected=2
	_, err = db.create(ctx, hook, client)
	require.NoError(t, err)

	// first delete sets connected=1
	_, err = db.delete(ctx, hook.id)
	require.Equal(t, errConnected, err)

	// second delete sets connected=0
	// should now succeed
	_, err = db.delete(ctx, hook.id)
	require.NoError(t, err)
}

func newTestDB(t *testing.T) *pgdb {
	return &pgdb{
		Database: sql.NewTestDB(t),
		factory: factory{
			Service:         fakeCloudService{},
			HostnameService: fakeHostnameService{},
		},
	}
}
