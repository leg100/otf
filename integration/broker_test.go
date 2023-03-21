package integration

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/pubsub"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBroker demonstrates publishing and subscribing of events via postgres.
func TestBroker(t *testing.T) {
	db := sql.NewTestDB(t)
	// local broker to which events are published
	local, err := pubsub.NewBroker(logr.Discard(), pubsub.BrokerConfig{PoolDB: db, PID: otf.String("123")})
	require.NoError(t, err)

	// remote broker from which events should be received
	remote, err := pubsub.NewBroker(logr.Discard(), pubsub.BrokerConfig{PoolDB: db, PID: otf.String("456")})
	require.NoError(t, err)
	// register org service with remote broker so that it knows how to
	// 'reassemble' org events
	remote.Register("organization", organization.NewTestService(t, db))

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() { cancel() })

	done := make(chan error)
	go func() {
		done <- local.Start(ctx)
	}()
	go func() {
		done <- remote.Start(ctx)
	}()

	t.Run("create organization event", func(t *testing.T) {
		localsub, err := local.Subscribe(ctx, "local-sub")
		require.NoError(t, err)
		remotesub, err := remote.Subscribe(ctx, "remote-sub")
		require.NoError(t, err)

		// sends event via local broker
		org := organization.CreateTestOrganization(t, db, organization.WithBroker(local))

		want := otf.Event{Type: otf.EventOrganizationCreated, Payload: org}
		// receive event on local broker
		assert.Equal(t, want, <-localsub)
		// receive event on remote broker (via postgres)
		assert.Equal(t, want, <-remotesub)
	})

	cancel()
	assert.NoError(t, <-done)
	assert.NoError(t, <-done)
}
