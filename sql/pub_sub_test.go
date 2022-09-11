package sql

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubSub(t *testing.T) {
	db := newTestDB(t)
	ps, err := NewPubSub(logr.Discard(), db.Pool())
	require.NoError(t, err)

	t.Run("cancel context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error)
		go func() {
			done <- ps.Start(ctx)
		}()
		cancel()
		assert.Equal(t, ctx.Err(), <-done)
	})
}

func TestPubSub_E2E(t *testing.T) {
	run := otf.NewTestRun(t, otf.TestRunCreateOptions{
		ID: otf.String("run-123"),
	})
	got := make(chan otf.Event, 1)
	ps := &PubSub{
		Logger:  logr.Discard(),
		pool:    newTestDB(t).Pool(),
		db:      &fakePubSubDB{run: run},
		local:   &fakePubSub{got: got},
		channel: "events_e2e_test",
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() { cancel() })

	go func() {
		ps.Start(ctx)
	}()

	// Give Start() some time to connect and start listening
	time.Sleep(100 * time.Millisecond)

	want := otf.Event{
		Type:    otf.EventRunStatusUpdate,
		Payload: run,
	}
	ps.Publish(want)

	// Give time for message to make its way via postgres and back.
	time.Sleep(time.Second)

	// Check that we only receive the one message forwarded locally, and not the
	// copy that is send via postgres as well
	assert.Equal(t, 1, len(got))
	assert.Equal(t, want, <-got)
}

func TestPubSub_marshal(t *testing.T) {
	event := otf.Event{
		Type: otf.EventRunStatusUpdate,
		Payload: otf.NewTestRun(t, otf.TestRunCreateOptions{
			ID: otf.String("run-123"),
		}),
	}

	got, err := marshal(event)
	require.NoError(t, err)
	want := fmt.Sprintf(
		"{\"relation\":\"run\",\"action\":\"status_update\",\"id\":\"run-123\",\"pid\":%d}",
		os.Getpid())
	assert.Equal(t, want, string(got))
}

func TestPubSub_reassemble(t *testing.T) {
	run := otf.NewTestRun(t, otf.TestRunCreateOptions{
		ID: otf.String("run-123"),
	})
	ps := PubSub{
		db: &fakePubSubDB{
			run: run,
		},
	}

	got, err := ps.reassemble(context.Background(), message{
		Table:  "run",
		Action: "status_update",
		ID:     "run-123",
	})
	require.NoError(t, err)
	want := otf.Event{
		Type:    otf.EventRunStatusUpdate,
		Payload: run,
	}
	assert.Equal(t, want, got)
}

type fakePubSubDB struct {
	run       *otf.Run
	workspace *otf.Workspace

	otf.DB
}

func (f *fakePubSubDB) GetRun(context.Context, string) (*otf.Run, error) {
	return f.run, nil
}

func (f *fakePubSubDB) GetWorkspace(context.Context, otf.WorkspaceSpec) (*otf.Workspace, error) {
	return f.workspace, nil
}

type fakePubSub struct {
	otf.PubSubService
	got chan otf.Event
}

func (f *fakePubSub) Publish(ev otf.Event) {
	f.got <- ev
}
