package agent

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/leg100/ots/mock"
	"github.com/stretchr/testify/assert"
)

// TestSupervisor_Start tests starting up the daemon and tests it handling a
// single job
func TestSupervisor_Start(t *testing.T) {
	want := &ots.Run{ID: "run-123", Status: tfe.RunPlanQueued}

	// Capture the run ID that is passed to the job processor
	got := make(chan string)

	supervisor := &Supervisor{
		Logger:       logr.Discard(),
		planRunnerFn: mockNewPlanRunnerFn,
		RunService: mock.RunService{
			UploadPlanLogsFn: func(id string, logs []byte, opts ots.AppendLogOptions) error {
				got <- id
				return nil
			},
		},
		Spooler:     newMockSpooler(want),
		concurrency: 1,
	}

	go supervisor.Start(context.Background())

	assert.Equal(t, "run-123", <-got)
}

// TestSupervisor_StartError tests starting up the agent daemon and tests it handling
// it a single job that errors
func TestSupervisor_StartError(t *testing.T) {
	// Mock run service and capture the run status it receives
	got := make(chan tfe.RunStatus)
	runService := &mock.RunService{
		UploadPlanLogsFn: func(id string, _ []byte, opts ots.AppendLogOptions) error { return nil },
		UpdateStatusFn: func(id string, status tfe.RunStatus) (*ots.Run, error) {
			got <- status
			return nil, nil
		},
	}

	supervisor := &Supervisor{
		Logger:       logr.Discard(),
		RunService:   runService,
		planRunnerFn: mockNewPlanRunnerFnWithError,
		Spooler: newMockSpooler(&ots.Run{
			ID:     "run-123",
			Status: tfe.RunPlanQueued,
		}),
		concurrency: 1,
	}

	go supervisor.Start(context.Background())

	// assert agent correctly propagates an errored status update
	assert.Equal(t, tfe.RunErrored, <-got)
}
