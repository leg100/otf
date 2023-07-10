package repo

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPurger_handler(t *testing.T) {
	ctx := context.Background()
	purger := newTestPurger(t)

	// add vcs provider
	err := purger.handle(ctx, pubsub.Event{
		Type:    pubsub.CreatedEvent,
		Payload: &vcsprovider.VCSProvider{ID: "vcs-1"},
	})
	require.NoError(t, err)

	assert.Len(t, purger.providers, 1)

	// add webhook referencing provider
	hook1 := &hook{id: uuid.New(), vcsProviderID: "vcs-1"}
	err = purger.handle(ctx, pubsub.Event{
		Type:    pubsub.CreatedEvent,
		Payload: hook1,
	})
	require.NoError(t, err)

	// add another webhook referencing same provider
	hook2 := &hook{id: uuid.New(), vcsProviderID: "vcs-1"}
	err = purger.handle(ctx, pubsub.Event{
		Type:    pubsub.CreatedEvent,
		Payload: hook2,
	})
	require.NoError(t, err)

	// should be 2 hooks, one provider
	assert.Len(t, purger.hooks, 2)
	assert.Len(t, purger.providers, 1)

	// delete first webhook
	err = purger.handle(ctx, pubsub.Event{
		Type:    pubsub.DeletedEvent,
		Payload: hook1,
	})
	require.NoError(t, err)

	// should be one hook, one provider
	assert.Len(t, purger.hooks, 1)
	assert.Len(t, purger.providers, 1)

	// delete second webhook
	err = purger.handle(ctx, pubsub.Event{
		Type:    pubsub.DeletedEvent,
		Payload: hook2,
	})
	require.NoError(t, err)

	// should be no hooks, no providers
	assert.Len(t, purger.hooks, 0)
	assert.Len(t, purger.providers, 0)

}

type (
	fakePurgerDB struct {
		purgerDB
	}
	fakeService struct {
		Service
	}
)

func newTestPurger(t *testing.T) *Purger {
	return &Purger{
		Logger: logr.Discard(),
		DB:     &fakePurgerDB{},
		cache:  newTestCache(t),
	}
}

func newTestCache(t *testing.T) *cache {
	cache, err := newCache(context.Background(), cacheOptions{
		VCSProviderService: &fakeVCSProviderService{},
		Service:            &fakeService{},
	})
	require.NoError(t, err)
	return cache
}

func (f *fakeService) ListWebhooks(ctx context.Context) ([]*hook, error) {
	return nil, nil
}

func (f *fakePurgerDB) Tx(context.Context, func(context.Context, pggen.Querier) error) error {
	return nil
}
