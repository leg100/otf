package agent

import (
	"errors"
	"fmt"

	"github.com/leg100/otf/internal"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
)

var ErrInvalidJobStateTransition = errors.New("invalid job state transition")

type JobStatus string

const (
	JobUnallocated JobStatus = "unallocated"
	JobAllocated   JobStatus = "allocated"
	JobRunning     JobStatus = "running"
	JobFinished    JobStatus = "finished"
	JobErrored     JobStatus = "errored"
	JobCanceled    JobStatus = "canceled"
)

type signal string

const (
	cancelSignal      signal = "cancel"
	forceCancelSignal signal = "force_cancel"
)

// JobSpec uniquely identifies a job.
type JobSpec struct {
	// ID of the run that this job is for.
	RunID string `json:"run_id"`
	// Phase of run that this job is for.
	Phase internal.PhaseType `json:"phase"`
}

type Job struct {
	JobSpec
	// Current status of job.
	Status JobStatus
	// Execution mode of job's workspace.
	ExecutionMode workspace.ExecutionMode
	// ID of job's workspace
	WorkspaceID string
	// ID of agent that this job is allocated to. Only set once job enters
	// JobAllocated state.
	AgentID *string
	// This indicates whether the run for the job has been signaled, i.e. user
	// has sent a cancelation request.
	signal *signal
}

func newJob(run *otfrun.Run) *Job {
	return &Job{
		JobSpec: JobSpec{
			RunID: run.ID,
			Phase: run.Phase(),
		},
		Status:        JobUnallocated,
		ExecutionMode: run.ExecutionMode,
		WorkspaceID:   run.WorkspaceID,
	}
}

func (j *Job) String() string { return fmt.Sprintf("%s-%s", j.RunID, j.Phase) }

func (j *Job) setSignal(s signal) error {
	if j.Status != JobRunning {
		return errors.New("job can only be signaled when in the JobRunning state")
	}
	j.signal = &s
	return nil
}

func (j *Job) allocate(agentID string) error {
	if j.Status != JobUnallocated {
		return errors.New("job can only be allocated when it is in the unallocated state")
	}
	j.Status = JobAllocated
	j.AgentID = &agentID
	return nil
}

func (j *Job) reallocate(agentID string) error {
	if j.Status != JobAllocated {
		return errors.New("job can only be re-allocated when it is in the allocated state")
	}
	j.AgentID = &agentID
	return nil
}

func (j *Job) updateStatus(to JobStatus) error {
	switch to {
	case JobRunning:
		if j.Status != JobAllocated {
			return ErrInvalidJobStateTransition
		}
	case JobFinished, JobErrored, JobCanceled:
		if j.Status != JobRunning {
			return ErrInvalidJobStateTransition
		}
	default:
		return ErrInvalidJobStateTransition
	}
	j.Status = to
	return nil
}
