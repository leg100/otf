package pubsub

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroker_Subscribe(t *testing.T) {
	broker := NewBroker(logr.Discard(), &fakePool{})

	sub, unsub := broker.Subscribe()

	assert.Equal(t, 1, len(broker.subs))

	cancel()
	<-sub
	assert.Equal(t, 0, len(broker.subs))
}

func TestBroker_UnsubscribeFullSubscriber(t *testing.T) {
	ctx := context.Background()
	broker := NewBroker(logr.Discard(), &fakePool{})

	_, err := broker.Subscribe(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, 1, len(broker.subs))

	// deliberating publish more than subBufferSize events to trigger broker to
	// unsubscribe the sub
	for i := 0; i < subBufferSize+1; i++ {
		broker.localPublish(Event{
			Type: EventType("test"),
		})
	}
	assert.Equal(t, 0, len(broker.subs))
}
