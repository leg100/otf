package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/leg100/ots/mock"
	"github.com/stretchr/testify/assert"
)

// TestAgentPoller tests a single iteration of the agent poller.
func TestAgentPoller(t *testing.T) {
	// Have the run service return a run when polled
	runService := &mock.RunService{
		GetQueuedFn: func(opts tfe.RunListOptions) (*ots.RunList, error) {
			return &ots.RunList{Items: []*ots.Run{{ID: "run-123", Status: tfe.RunPlanQueued}}}, nil
		},
	}

	// Mock the processor and capture the run ID that is passed
	got := make(chan string)
	processor := mockProcessor{
		PlanFn: func(ctx context.Context, run *ots.Run, path string) error {
			got <- run.ID
			return nil
		},
	}

	agent := &Agent{
		ConfigurationVersionService: &mock.ConfigurationVersionService{},
		StateVersionService:         &mock.StateVersionService{},
		PlanService:                 &mock.PlanService{},
		RunService:                  runService,
		Processor:                   &processor,
	}

	ctx, cancel := context.WithCancel(context.Background())

	go agent.Poller(ctx)

	assert.Equal(t, "run-123", <-got)

	cancel()
}

// Test poller error handling. The processor returns an error and the poller
// should update the plan with an error status.
func TestAgentPollerError(t *testing.T) {
	// Mock run service and capture the plan status it receives
	got := make(chan tfe.PlanStatus)
	runService := &mock.RunService{
		GetQueuedFn: func(opts tfe.RunListOptions) (*ots.RunList, error) {
			return &ots.RunList{Items: []*ots.Run{
				{
					ID: "run-123",
					Plan: &ots.Plan{
						ExternalID: "plan-123",
					},
					Status: tfe.RunPlanQueued,
				},
			}}, nil
		},
		UpdatePlanStatusFn: func(id string, status tfe.PlanStatus) (*ots.Run, error) {
			got <- status
			return nil, nil
		},
	}

	// Mock processor returning an error
	processor := mockProcessor{
		PlanFn: func(ctx context.Context, run *ots.Run, path string) error {
			return errors.New("mock process error")
		},
	}

	agent := &Agent{
		ConfigurationVersionService: &mock.ConfigurationVersionService{},
		StateVersionService:         &mock.StateVersionService{},
		PlanService:                 &mock.PlanService{},
		RunService:                  runService,
		Processor:                   &processor,
	}

	ctx, cancel := context.WithCancel(context.Background())

	go agent.Poller(ctx)

	assert.Equal(t, tfe.PlanErrored, <-got)

	cancel()
}
