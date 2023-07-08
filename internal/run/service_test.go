package run

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Watch(t *testing.T) {
	// input event channel
	in := make(chan pubsub.Event, 1)

	svc := &service{
		site:          internal.NewAllowAllAuthorizer(),
		Logger:        logr.Discard(),
		PubSubService: &fakeSubscriber{ch: in},
	}

	// inject input event
	want := pubsub.Event{
		Payload: &Run{},
	}
	in <- want

	got, err := svc.Watch(context.Background(), WatchOptions{})
	require.NoError(t, err)

	assert.Equal(t, want, <-got)
}
