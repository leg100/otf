package agent

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
	"github.com/leg100/otf/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRunLister struct {
	runs []*otf.Run
}

func (l *mockRunLister) List(opts otf.RunListOptions) (*otf.RunList, error) {
	return &otf.RunList{Items: l.runs}, nil
}

type mockSubscription struct {
	c chan otf.Event
}

func (s *mockSubscription) C() <-chan otf.Event { return s.c }

func (s *mockSubscription) Close() error { return nil }

// TestSpooler_New tests the spooler constructor
func TestSpooler_New(t *testing.T) {
	want := &otf.Run{ID: "run-123", Status: tfe.RunPlanQueued}

	spooler, err := NewSpooler(
		&mockRunLister{runs: []*otf.Run{want}},
		&mock.EventService{},
		logr.Discard(),
	)
	require.NoError(t, err)

	assert.Equal(t, want, <-spooler.queue)
}

// TestSpooler_Start tests the spooler daemon start op
func TestSpooler_Start(t *testing.T) {
	spooler := &SpoolerDaemon{
		EventService: &mock.EventService{
			SubscribeFn: func(id string) otf.Subscription {
				return &mockSubscription{}
			},
		},
		Logger: logr.Discard(),
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		spooler.Start(ctx)
		done <- struct{}{}
	}()

	cancel()

	<-done
}

// TestSpooler_GetJob tests retrieving a job from the spooler with a
// pre-populated queue
func TestSpooler_GetJob(t *testing.T) {
	want := &otf.Run{ID: "run-123", Status: tfe.RunPlanQueued}

	spooler := &SpoolerDaemon{queue: make(chan otf.Job, 1)}
	spooler.queue <- want

	assert.Equal(t, want, <-spooler.GetJob())
}

// TestSpooler_GetJobFromEvent tests retrieving a job from the spooler after an
// event is received
func TestSpooler_GetJobFromEvent(t *testing.T) {
	want := &otf.Run{ID: "run-123", Status: tfe.RunPlanQueued}

	sub := mockSubscription{c: make(chan otf.Event, 1)}

	spooler := &SpoolerDaemon{
		queue: make(chan otf.Job, 1),
		EventService: &mock.EventService{
			SubscribeFn: func(id string) otf.Subscription {
				return &sub
			},
		},
		Logger: logr.Discard(),
	}

	go spooler.Start(context.Background())

	// send event
	sub.c <- otf.Event{Type: otf.PlanQueued, Payload: want}

	assert.Equal(t, want, <-spooler.GetJob())
}

// TestSpooler_GetJobFromCancelation tests retrieving a job from the spooler
// after a cancelation is received
func TestSpooler_GetJobFromCancelation(t *testing.T) {
	want := &otf.Run{ID: "run-123", Status: tfe.RunCanceled}

	sub := mockSubscription{c: make(chan otf.Event, 1)}

	spooler := &SpoolerDaemon{
		cancelations: make(chan otf.Job, 1),
		EventService: &mock.EventService{
			SubscribeFn: func(id string) otf.Subscription {
				return &sub
			},
		},
		Logger: logr.Discard(),
	}

	go spooler.Start(context.Background())

	// send event
	sub.c <- otf.Event{Type: otf.RunCanceled, Payload: want}

	assert.Equal(t, want, <-spooler.GetCancelation())
}
