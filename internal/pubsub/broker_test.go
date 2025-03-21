package pubsub

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
)

type foo struct {
	id resource.TfeID
}

func fooGetter(ctx context.Context, id resource.TfeID, action sql.Action) (*foo, error) {
	return &foo{id: id}, nil
}

func TestBroker_Subscribe(t *testing.T) {
	ctx := context.Background()
	broker := NewBroker[*foo](logr.Discard(), &fakeListener{}, "foos", nil)

	sub, unsub := broker.Subscribe(ctx)
	assert.Equal(t, 1, len(broker.subs))

	unsub()
	<-sub
	assert.Equal(t, 0, len(broker.subs))
}

func TestBroker_UnsubscribeViaContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	broker := NewBroker[*foo](logr.Discard(), &fakeListener{}, "foos", nil)

	sub, _ := broker.Subscribe(ctx)
	assert.Equal(t, 1, len(broker.subs))

	cancel()
	<-sub
	assert.Equal(t, 0, len(broker.subs))
}

func TestBroker_forward(t *testing.T) {
	ctx := context.Background()
	broker := NewBroker(logr.Discard(), &fakeListener{}, "foos", fooGetter)

	sub, unsub := broker.Subscribe(ctx)
	defer unsub()

	broker.forward(ctx, "foo-bar", sql.InsertAction)
	want := Event[*foo]{
		Type:    CreatedEvent,
		Payload: &foo{id: testutils.ParseID(t, "foo-bar")},
	}
	assert.Equal(t, want, <-sub)
}

func TestBroker_UnsubscribeFullSubscriber(t *testing.T) {
	ctx := context.Background()
	broker := NewBroker(logr.Discard(), &fakeListener{}, "foos", fooGetter)

	broker.Subscribe(ctx)
	assert.Equal(t, 1, len(broker.subs))

	// deliberating publish more than subBufferSize events to trigger broker to
	// unsubscribe the sub
	for i := 0; i < subBufferSize+1; i++ {
		broker.forward(ctx, "foo-123", sql.InsertAction)
	}
	assert.Equal(t, 0, len(broker.subs))
}
