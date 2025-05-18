package pubsub

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
)

type foo struct {
	Bar string `json:"bar"`
}

func TestBroker_Subscribe(t *testing.T) {
	ctx := context.Background()
	broker := NewBroker[*foo](logr.Discard(), &fakeListener{}, "foos")

	sub, unsub := broker.Subscribe(ctx)
	assert.Equal(t, 1, len(broker.subs))

	unsub()
	<-sub
	assert.Equal(t, 0, len(broker.subs))
}

func TestBroker_UnsubscribeViaContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	broker := NewBroker[*foo](logr.Discard(), &fakeListener{}, "foos")

	sub, _ := broker.Subscribe(ctx)
	assert.Equal(t, 1, len(broker.subs))

	cancel()
	<-sub
	assert.Equal(t, 0, len(broker.subs))
}

func TestBroker_forward(t *testing.T) {
	ctx := context.Background()
	broker := NewBroker[*foo](logr.Discard(), &fakeListener{}, "foos")

	sub, unsub := broker.Subscribe(ctx)
	defer unsub()

	broker.forward(sql.InsertAction, base64.StdEncoding.EncodeToString([]byte(`{"bar": "baz"}`)))
	want := Event[*foo]{
		Type:    CreatedEvent,
		Payload: &foo{Bar: "baz"},
	}
	assert.Equal(t, want, <-sub)
}

func TestBroker_UnsubscribeFullSubscriber(t *testing.T) {
	ctx := context.Background()
	broker := NewBroker[*foo](logr.Discard(), &fakeListener{}, "foos")

	broker.Subscribe(ctx)
	assert.Equal(t, 1, len(broker.subs))

	// deliberating publish more than subBufferSize events to trigger broker to
	// unsubscribe the sub
	for range subBufferSize + 1 {
		broker.forward(sql.InsertAction, base64.StdEncoding.EncodeToString([]byte(`{"bar": "baz"}`)))
	}
	assert.Equal(t, 0, len(broker.subs))
}
