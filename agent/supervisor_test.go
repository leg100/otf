package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/ots"
	agentmock "github.com/leg100/ots/agent/mock"
	"github.com/leg100/ots/mock"
	"github.com/stretchr/testify/assert"
)

// TestSupervisor_Start tests starting up the daemon and tests it handling a
// single job
func TestSupervisor_Start(t *testing.T) {
	want := &agentmock.Job{
		ID:     "run-123",
		Status: "queued",
		DoFn: func(env *ots.Environment) error {
			return nil
		},
	}

	// Capture the run ID and status that is received upon finishing
	got := make(chan string)

	supervisor := &Supervisor{
		Logger: logr.Discard(),
		RunService: mock.RunService{
			StartFn: func(id string, opts ots.JobStartOptions) (ots.Job, error) {
				got <- id
				return want, nil
			},
			FinishFn: func(id string, opts ots.JobFinishOptions) (ots.Job, error) {
				got <- id
				return want, nil
			},
			UploadLogsFn: func(id string, logs []byte, opts ots.PutChunkOptions) error {
				got <- id
				return nil
			},
		},
		JobGetter:   NewMockJobGetter(want),
		concurrency: 1,
	}

	go supervisor.Start(context.Background())

	assert.Equal(t, "run-123", <-got)
}

// TestSupervisor_StartError tests starting up the agent daemon and tests it handling
// it a single job that errors
func TestSupervisor_StartError(t *testing.T) {
	want := &agentmock.Job{
		ID:     "run-123",
		Status: "queued",
		DoFn: func(env *ots.Environment) error {
			return errors.New("mock error")
		},
	}

	// Capture whether the job finishes with an error or not.
	got := make(chan bool)

	supervisor := &Supervisor{
		Logger: logr.Discard(),
		RunService: mock.RunService{
			StartFn:      func(id string, opts ots.JobStartOptions) (ots.Job, error) { return want, nil },
			UploadLogsFn: func(id string, logs []byte, opts ots.PutChunkOptions) error { return nil },
			FinishFn: func(id string, opts ots.JobFinishOptions) (ots.Job, error) {
				got <- opts.Errored
				return want, nil
			},
		},
		JobGetter:   NewMockJobGetter(want),
		concurrency: 1,
	}

	go supervisor.Start(context.Background())

	// Finish should propagate an error
	assert.True(t, <-got)
}
