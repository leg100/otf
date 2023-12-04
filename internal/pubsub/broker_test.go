package pubsub

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
)

type foo struct {
	id string
}

func fooGetter(ctx context.Context, id string, action sql.Action) (*foo, error) {
	return &foo{id: id}, nil
}

func TestBroker_Subscribe(t *testing.T) {
	broker := NewBroker[*foo](logr.Discard(), &fakeListener{}, "foos", nil)

	sub, unsub := broker.Subscribe()
	assert.Equal(t, 1, len(broker.subs))

	unsub()
	<-sub
	assert.Equal(t, 0, len(broker.subs))
}

func TestBroker_forward(t *testing.T) {
	broker := NewBroker[*foo](logr.Discard(), &fakeListener{}, "foos", fooGetter)

	sub, unsub := broker.Subscribe()
	defer unsub()

	ctx := context.Background()
	broker.forward(ctx, "bar", sql.InsertAction)
	want := Event[*foo]{
		Type:    CreatedEvent,
		Payload: &foo{id: "bar"},
	}
	assert.Equal(t, want, <-sub)
}

func TestBroker_UnsubscribeFullSubscriber(t *testing.T) {
	ctx := context.Background()
	broker := NewBroker[*foo](logr.Discard(), &fakeListener{}, "foos", fooGetter)

	broker.Subscribe()
	assert.Equal(t, 1, len(broker.subs))

	// deliberating publish more than subBufferSize events to trigger broker to
	// unsubscribe the sub
	for i := 0; i < subBufferSize+1; i++ {
		broker.forward(ctx, "bar", sql.InsertAction)
	}
	assert.Equal(t, 0, len(broker.subs))
}
