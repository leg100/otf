package otf

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrJobAlreadyClaimed = errors.New("job already claimed")
)

// Job is a piece of work to do.
type Job interface {
	// Do does the piece of work in an execution environment
	Do(Environment) error
	// GetID gets the ID of the Job
	GetID() string
	// GetService gets the appropriate service for interacting with the job
	GetService(Application) JobService
}

type JobService interface {
	// Claim claims a job entitling the caller to do the job.
	// ErrJobAlreadyClaimed is returned if job is already claimed.
	Claim(ctx context.Context, id string, opts JobClaimOptions) (Job, error)
	// Finish is called by an agent when it finishes a job.
	Finish(ctx context.Context, id string, opts JobFinishOptions) (Job, error)
	// PutChunk uploads a chunk of logs from the job.
	PutChunk(ctx context.Context, id string, chunk Chunk) error
}

type JobClaimOptions struct {
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
	switch run.Status() {
	case RunPlanQueued, RunPlanning:
		return run.Plan, jsp.PlanService, nil
	case RunApplyQueued, RunApplying:
		return run.Apply, jsp.ApplyService, nil
	default:
		return nil, nil, fmt.Errorf("attempted to retrieve active job for run but run as an invalid status: %s", run.Status())
	}
}
