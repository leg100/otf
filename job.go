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
	// GetStatus gets the status of the Job
	GetStatus() string
}

type JobService interface {
	// List retrieves pending jobs
	List(ctx context.Context) ([]Job, error)
	// Claim claims a job entitling the caller to do the job.
	// ErrJobAlreadyClaimed is returned if job is already claimed.
	Claim(ctx context.Context, id string, opts JobClaimOptions) error
	// Finish is called by an agent when it finishes a job.
	Finish(ctx context.Context, id string, opts JobFinishOptions) (*Run, error)
	// ChunkStore handles reading and writing chunks of logs for jobs
	ChunkStore
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
	switch run.Status {
	case RunPlanQueued, RunPlanning:
		return run.Plan, jsp.PlanService, nil
	case RunApplyQueued, RunApplying:
		return run.Apply, jsp.ApplyService, nil
	default:
		return nil, nil, fmt.Errorf("attempted to retrieve active job for run but run as an invalid status: %s", run.Status)
	}
}
