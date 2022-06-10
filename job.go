package otf

import (
	"context"
	"errors"
)

var (
	ErrJobAlreadyClaimed = errors.New("job already claimed")
)

// Job is a piece of work to do.
type Job interface {
	// Do does the piece of work in an execution environment
	Do(Environment) error
	// GetID gets the ID of the Job
	JobID() string
}

type JobService interface {
	// Claim claims a job entitling the caller to do the job.
	// ErrJobAlreadyClaimed is returned if job is already claimed.
	Claim(ctx context.Context, id string, opts JobClaimOptions) (Job, error)
	// Finish is called by an agent when it finishes a job.
	Finish(ctx context.Context, id string, opts JobFinishOptions) (Job, error)
	// Retrieve and upload chunks of logs for jobs
	ChunkService
}

type JobClaimOptions struct {
	AgentID string
}

type JobFinishOptions struct {
	Errored bool
}
