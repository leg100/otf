package pubsub

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
)

func TestBroker_Subscribe(t *testing.T) {
	type foo struct{}

	broker := NewBroker[*foo](logr.Discard(), &fakeListener{}, "foos", nil)

	sub, unsub := broker.Subscribe()
	assert.Equal(t, 1, len(broker.subs))

	unsub()
	<-sub
	assert.Equal(t, 0, len(broker.subs))
}

func TestBroker_UnsubscribeFullSubscriber(t *testing.T) {
	type foo struct {
		id string
	}
	getter := func(ctx context.Context, id string, action sql.Action) (*foo, error) {
		return &foo{id: id}, nil
	}

	ctx := context.Background()
	broker := NewBroker[*foo](logr.Discard(), &fakeListener{}, "foos", getter)

	broker.Subscribe()
	assert.Equal(t, 1, len(broker.subs))

	// deliberating publish more than subBufferSize events to trigger broker to
	// unsubscribe the sub
	for i := 0; i < subBufferSize+1; i++ {
		broker.forward(ctx, "bar", sql.InsertAction)
	}
	assert.Equal(t, 0, len(broker.subs))
}
