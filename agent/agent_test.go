package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/leg100/ots/mock"
	"github.com/stretchr/testify/assert"
)

// TestAgent_Start tests starting up the agent daemon and tests it handling a
// single job
func TestAgent_Start(t *testing.T) {
	want := &ots.Run{ID: "run-123", Status: tfe.RunPlanQueued}

	// Capture the run ID that is passed to the job processor
	got := make(chan string)

	agent := &Agent{
		Logger: logr.Discard(),
		Processor: &mockProcessor{
			PlanFn: func(ctx context.Context, run *ots.Run, path string) error {
				got <- run.ID
				return nil
			},
		},
		Spooler: newMockSpooler(want),
	}

	go agent.Start(context.Background())

	assert.Equal(t, "run-123", <-got)
}

// TestAgent_StartError tests starting up the agent daemon and tests it handling
// it a single job that errors
func TestAgent_StartError(t *testing.T) {
	// Mock run service and capture the plan status it receives
	got := make(chan tfe.PlanStatus)
	runService := &mock.RunService{
		UpdatePlanStatusFn: func(id string, status tfe.PlanStatus) (*ots.Run, error) {
			got <- status
			return nil, nil
		},
	}

	// Mock job returning an error
	processor := mockProcessor{
		PlanFn: func(ctx context.Context, run *ots.Run, path string) error {
			return errors.New("mock process error")
		},
	}

	agent := &Agent{
		Logger:     logr.Discard(),
		RunService: runService,
		Processor:  &processor,
		Spooler: newMockSpooler(&ots.Run{
			ID:     "run-123",
			Status: tfe.RunPlanQueued,
		}),
	}

	go agent.Start(context.Background())

	// assert agent correctly propagates a plan errored status update
	assert.Equal(t, tfe.PlanErrored, <-got)
}
