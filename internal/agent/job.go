package agent

import (
	"fmt"

	"github.com/leg100/otf/internal"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
)

type JobStatus string

const (
	JobUnallocated JobStatus = "unallocated"
	JobAllocated   JobStatus = "allocated"
	JobRunning     JobStatus = "running"
	JobFinished    JobStatus = "finished"
	JobErrored     JobStatus = "errored"
	JobCanceled    JobStatus = "canceled"
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
