package pubsub

import (
	"context"
	"testing"

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
