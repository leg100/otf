package pubsub

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubSub(t *testing.T) {
	ps := newSpoke()
	ev := otf.Event{Type: "eclipse"}

	// create subscribers
	ctx1, cancel1 := context.WithCancel(context.Background())
	sub1, err := ps.Subscribe(ctx1, "sub1")
	require.NoError(t, err)

	ctx2, cancel2 := context.WithCancel(context.Background())
	sub2, err := ps.Subscribe(ctx2, "sub2")
	require.NoError(t, err)

	assert.Equal(t, 2, len(ps.subs))

	// publish event and check subscribers both receive it
	ps.Publish(ev)

	assert.Equal(t, ev, <-sub1)
	assert.Equal(t, ev, <-sub2)

	// unsubscribe subscribers - ctx cancellation is a concurrent op so we need
	// to give it some time to take affect before checking
	cancel2()
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, len(ps.subs))

	cancel1()
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, len(ps.subs))
}
