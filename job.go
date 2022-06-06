package otf

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	JobCanceled      JobStatus = "canceled"
	JobForceCanceled JobStatus = "force_canceled"
	JobErrored       JobStatus = "errored"
	JobPending       JobStatus = "pending"
	JobClaimed       JobStatus = "claimed"
	JobRunning       JobStatus = "running"
	JobFinished      JobStatus = "finished"
)

var (
	ErrJobAlreadyClaimed = errors.New("job already claimed")
)

// Doer does some work.
type Doer interface {
	// Do some work in an execution environment
	Do(Environment) error
}

type JobStatus string

// Job is a unit of work to do.
type Job struct {
	id        string
	createdAt time.Time
	status    JobStatus
	Doer
}

func NewJob(id string, doer Doer) *Job {
	return &Job{
		id:        NewID("job"),
		createdAt: CurrentTimestamp(),
		status:    JobPending,
		Doer:      doer,
	}
}

type JobService interface {
	// Claim claims a job entitling the caller to do the job.
	// ErrJobAlreadyClaimed is returned if job is already claimed.
	Claim(ctx context.Context, id string, opts JobClaimOptions) (*Job, error)
	// Finish is called by an agent when it finishes a job.
	Finish(ctx context.Context, id string, opts JobFinishOptions) (*Job, error)
	// PutChunk uploads a chunk of logs from the job.
	PutChunk(ctx context.Context, id string, chunk Chunk) error
}

type JobClaimOptions struct {
	AgentID string
}

type JobFinishOptions struct {
	Errored bool
}

// JobStore persists jobs
type JobStore interface {
	Create(ctx context.Context, job *Job) error
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
		return nil, nil, fmt.Errorf("attempted to retrieve active job for run but run has an invalid status: %s", run.Status())
	}
}

type PlanJob struct {
	// for retrieving config tarball
	configurationVersionID string
	// for retrieving latest state
	workspaceID string
	// for uploading plan file
	planID string
	// flags for terraform plan
	isDestroy bool
}

type ApplyJob struct {
	// for retrieving config tarball
	configurationVersionID string
	// for retrieving latest state and creating new state
	workspaceID string
	// for retrieving plan file
	runID string
	// flags for terraform apply
	isDestroy bool
}
