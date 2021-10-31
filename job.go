package otf

import (
	"context"
	"fmt"
)

type ErrJobAlreadyStarted error

// Job is either a Run's Plan or Apply.
type Job interface {
	// Do does the piece of work in an execution environment
	Do(*Run, Environment) error

	// GetID gets the ID of the Job
	GetID() string

	// GetStatus gets the status of the Job
	GetStatus() string
}

type JobService interface {
	// Start is called by an agent when it starts a job. ErrJobAlreadyStarted
	// should be returned if another agent has already started it.
	Start(ctx context.Context, id string, opts JobStartOptions) (*Run, error)

	// Finish is called by an agent when it finishes a job.
	Finish(ctx context.Context, id string, opts JobFinishOptions) (*Run, error)

	// ChunkStore handles putting and getting chunks of logs
	ChunkStore
}

type JobStartOptions struct {
	AgentID string
}

type JobFinishOptions struct {
	Errored bool
}

// JobSelector selects the appropriate job and job service for a Run
type JobSelector struct {
	PlanService  PlanService
	ApplyService ApplyService
}

// GetJob returns the appropriate job and job service for the Run
func (jsp *JobSelector) GetJob(run *Run) (Job, JobService, error) {
	switch run.Status {
	case RunPlanQueued, RunPlanning:
		return run.Plan, jsp.PlanService, nil
	case RunApplyQueued, RunApplying:
		return run.Apply, jsp.ApplyService, nil
	default:
		return nil, nil, fmt.Errorf("attempted to retrieve active job for run but run as an invalid status: %s", run.Status)
	}
}
