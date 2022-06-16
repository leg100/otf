package otf

import (
	"context"
	"errors"
	"time"
)

const (
	JobPending     JobStatus = "pending"
	JobQueued      JobStatus = "queued"
	JobRunning     JobStatus = "running"
	JobFinished    JobStatus = "finished"
	JobCanceled    JobStatus = "canceled"
	JobErrored     JobStatus = "errored"
	JobUnreachable JobStatus = "unreachable"
)

var (
	ErrJobAlreadyClaimed = errors.New("job already claimed")
)

type JobStatus string

// Job is a unit of work
type Job interface {
	// Do some work in an execution environment
	Do(Environment) error
	// GetID gets the ID of the Job
	JobID() string
	JobStatus() JobStatus
	// Get job status timestamps
	JobStatusTimestamps() []JobStatusTimestamp
	// Lookup timestamp for a job status
	JobStatusTimestamp(status JobStatus) (time.Time, error)
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

type JobStatusTimestamp struct {
	Status    JobStatus
	Timestamp time.Time
}

type job struct {
	id               string
	status           JobStatus
	statusTimestamps []JobStatusTimestamp
}

func (j *job) JobID() string                          { return j.id }
func (j *job) Status() JobStatus                      { return j.status }
func (j *job) StatusTimestamps() []JobStatusTimestamp { return j.statusTimestamps }

func (j *job) StatusTimestamp(status JobStatus) (time.Time, error) {
	for _, pst := range j.statusTimestamps {
		if pst.Status == status {
			return pst.Timestamp, nil
		}
	}
	return time.Time{}, ErrStatusTimestampNotFound
}

func (j *job) updateStatus(status JobStatus) {
	j.status = status
	j.statusTimestamps = append(j.statusTimestamps, JobStatusTimestamp{
		Status:    status,
		Timestamp: CurrentTimestamp(),
	})
}

func newJob() *job {
	return &job{
		status: JobPending,
		id:     NewID("job"),
	}
}
