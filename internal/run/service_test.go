package run

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	// input event channel
	in := make(chan internal.Event, 1)

	svc := &service{
		site:          internal.NewAllowAllAuthorizer(),
		Logger:        logr.Discard(),
		PubSubService: &fakeSubscriber{ch: in},
	}

	// inject input event
	want := internal.Event{
		Payload: &Run{},
		Type:    internal.EventRunCreated,
	}
	in <- want

	got, err := svc.Watch(context.Background(), WatchOptions{})
	require.NoError(t, err)

	assert.Equal(t, want, <-got)
}
