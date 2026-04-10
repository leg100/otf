package pubsub

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type foo struct {
	Bar string `json:"bar"`
}

func TestBroker_Subscribe(t *testing.T) {
	broker := NewBroker[*foo](logr.Discard(), "foos")

	sub, unsub, err := broker.Subscribe(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 1, len(broker.subs))

	unsub()
	<-sub
	assert.Equal(t, 0, len(broker.subs))
}

func TestBroker_UnsubscribeViaContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	broker := NewBroker[*foo](logr.Discard(), "foos")

	sub, _, err := broker.Subscribe(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(broker.subs))

	cancel()
	<-sub
	assert.Equal(t, 0, len(broker.subs))
}

func TestBroker_forward(t *testing.T) {
	broker := NewBroker[*foo](logr.Discard(), "foos")

	sub, unsub, err := broker.Subscribe(t.Context())
	require.NoError(t, err)
	defer unsub()

	broker.Forward(sql.Event{Action: sql.InsertAction, Record: []byte(`{"bar": "baz"}`)})
	want := Event[*foo]{
		Type:    CreatedEvent,
		Payload: &foo{Bar: "baz"},
	}
	assert.Equal(t, want, <-sub)
}

func TestBroker_UnsubscribeFullSubscriber(t *testing.T) {
	broker := NewBroker[*foo](logr.Discard(), "foos")

	broker.Subscribe(t.Context())
	assert.Equal(t, 1, len(broker.subs))

	// deliberating publish more than subBufferSize events to trigger broker to
	// unsubscribe the sub
	for range subBufferSize + 1 {
		broker.Forward(sql.Event{Action: sql.InsertAction, Record: []byte(`{"bar": "baz"}`)})
	}
	assert.Equal(t, 0, len(broker.subs))
}
