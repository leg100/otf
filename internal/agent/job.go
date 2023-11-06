package agent

import (
	"fmt"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/workspace"
)

type JobStatus string

const (
	JobUnallocated JobStatus = "unallocated"
	JobAllocated   JobStatus = "allocated"
	JobRunning     JobStatus = "running"
	JobFinished    JobStatus = "finished"
	JobErrored     JobStatus = "errored"
	// JobCanceled?
)

type Job struct {
	JobSpec
	// Current status of job.
	Status JobStatus
	// ID of agent that this job is allocated to.
	AgentID string
	// Execution mode of job's workspace.
	ExecutionMode workspace.ExecutionMode
	// ID of job's workspace
	WorkspaceID string
}

func (j *Job) String() string { return fmt.Sprintf("%s-%s", j.RunID, j.Phase) }

// JobSpec uniquely identifies a job.
type JobSpec struct {
	// ID of the run that this job is for.
	RunID string `json:"run_id"`
	// Phase of run that this job is for.
	Phase internal.PhaseType `json:"phase"`
}

//func newJob(run *otfrun.Run, agentID string) *Job {
//	return &Job{
//		RunID:   run.ID,
//		Phase:   run.Phase(),
//		Status:  JobPending,
//		AgentID: agentID,
//	}
//}
