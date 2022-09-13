package sql

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPubSub_E2E tests that one pubsub process can publish a message and that
// another pubsub process can receive it.
func TestPubSub_E2E(t *testing.T) {
	run := otf.NewTestRun(t, otf.TestRunCreateOptions{
		ID: otf.String("run-123"),
	})
	senderGot := make(chan otf.Event, 1)
	sender := &PubSub{
		Logger:  logr.Discard(),
		pool:    newTestDB(t).Pool(),
		db:      &fakePubSubDB{run: run},
		local:   &fakePubSub{got: senderGot},
		channel: "events_e2e_test",
		pid:     "sender-1",
	}
	// record what receiver receives
	receiverGot := make(chan otf.Event, 1)
	receiver := &PubSub{
		Logger:  logr.Discard(),
		pool:    newTestDB(t).Pool(),
		db:      &fakePubSubDB{run: run},
		local:   &fakePubSub{got: receiverGot},
		channel: "events_e2e_test",
		pid:     "receiver-1",
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() { cancel() })

	go func() { sender.Start(ctx) }()
	go func() { receiver.Start(ctx) }()

	// Give Start() time to connect and start listening
	time.Sleep(100 * time.Millisecond)

	want := otf.Event{
		Type:    otf.EventRunStatusUpdate,
		Payload: run,
	}
	sender.Publish(want)

	// Give time for message to make its way via postgres and back.
	time.Sleep(time.Second)

	// We expect the receiver process to have received a copy
	assert.Equal(t, 1, len(receiverGot))
	assert.Equal(t, want, <-receiverGot)

	// We also expect the sender process to have published a copy locally for local
	// subs
	assert.Equal(t, 1, len(senderGot))
	assert.Equal(t, want, <-senderGot)
}

func TestPubSub_marshal(t *testing.T) {
	ps := &PubSub{pid: "process-1"}
	event := otf.Event{
		Type: otf.EventRunStatusUpdate,
		Payload: otf.NewTestRun(t, otf.TestRunCreateOptions{
			ID: otf.String("run-123"),
		}),
	}

	got, err := ps.marshal(event)
	require.NoError(t, err)
	want := "{\"relation\":\"run\",\"action\":\"status_update\",\"id\":\"run-123\",\"pid\":\"process-1\"}"
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
