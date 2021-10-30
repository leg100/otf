package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	agentmock "github.com/leg100/otf/agent/mock"
	"github.com/leg100/otf/mock"
	"github.com/stretchr/testify/assert"
)

// TestSupervisor_Start tests starting up the supervisor and tests it handling a
// single run
func TestSupervisor_Start(t *testing.T) {
	want := &agentmock.Job{
		ID:     "run-123",
		Status: "queued",
		DoFn: func(env *otf.Executor) error {
			return nil
		},
	}

	// Capture the run ID and status that is received upon finishing
	got := make(chan string)

	supervisor := &Supervisor{
		Logger: logr.Discard(),
		RunService: mock.RunService{
			StartFn: func(id string, opts otf.JobStartOptions) (otf.Job, error) {
				got <- id
				return want, nil
			},
			FinishFn: func(id string, opts otf.JobFinishOptions) (otf.Job, error) {
				got <- id
				return want, nil
			},
			UploadLogsFn: func(ctx context.Context, id string, logs []byte, opts otf.RunUploadLogsOptions) error {
				got <- id
				return nil
			},
		},
		Spooler:     NewMockSpooler(WithMockJobs(want)),
		concurrency: 1,
		Terminator:  NewTerminator(),
	}

	go supervisor.Start(context.Background())

	assert.Equal(t, "run-123", <-got)
}

// TestSupervisor_StartError tests starting up the supervisor and tests it
// handling it a single job that errors
func TestSupervisor_StartError(t *testing.T) {
	want := &agentmock.Job{
		ID:     "run-123",
		Status: "queued",
		DoFn: func(env *otf.Executor) error {
			return errors.New("mock error")
		},
	}

	// Capture whether the job finishes with an error or not.
	got := make(chan bool)

	supervisor := &Supervisor{
		Logger: logr.Discard(),
		RunService: mock.RunService{
			StartFn:      func(id string, opts otf.JobStartOptions) (otf.Job, error) { return want, nil },
			UploadLogsFn: func(ctx context.Context, id string, logs []byte, opts otf.RunUploadLogsOptions) error { return nil },
			FinishFn: func(id string, opts otf.JobFinishOptions) (otf.Job, error) {
				got <- opts.Errored
				return want, nil
			},
		},
		Spooler:     NewMockSpooler(WithMockJobs(want)),
		concurrency: 1,
		Terminator:  NewTerminator(),
	}

	go supervisor.Start(context.Background())

	// Finish should propagate an error
	assert.True(t, <-got)
}
