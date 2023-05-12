package pubsub

import (
	"context"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroker_Subscribe(t *testing.T) {
	broker := NewBroker(logr.Discard(), &fakePool{})
	ctx, cancel := context.WithCancel(context.Background())

	sub, err := broker.Subscribe(ctx, "")
	require.NoError(t, err)

	assert.Equal(t, 1, len(broker.subs))

	cancel()
	<-sub
	assert.Equal(t, 0, len(broker.subs))
}

func TestBroker_Publish(t *testing.T) {
	ctx := context.Background()
	broker := NewBroker(logr.Discard(), &fakePool{})

	sub, err := broker.Subscribe(ctx, "")
	require.NoError(t, err)

	event := internal.Event{
		Type: internal.EventType("payload_update"),
	}
	broker.Publish(event)

	assert.Equal(t, event, <-sub)
}

func TestPubSub_receive(t *testing.T) {
	ctx := context.Background()
	broker := NewBroker(logr.Discard(), &fakePool{})
	sub, err := broker.Subscribe(ctx, "")
	require.NoError(t, err)

	// Fake the payload to be returned from the event unmarshaler
	payload := struct{ ID string }{ID: "run-123"}
	broker.Register("runs", &fakeUnmarshaler{resource: payload})

	// This is the fake notification that would normally be received from postgres
	err = broker.receive(ctx, &pgconn.Notification{
		Payload: `{"table":"runs","op":"UPDATE","record":{"id": "run-123"}}`,
	})
	require.NoError(t, err)

	// expect the given event to be returned.
	want := internal.Event{
		Type:    internal.UpdatedEvent,
		Payload: payload,
	}
	assert.Equal(t, want, <-sub)
}
