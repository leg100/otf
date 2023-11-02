package agent

import (
	"github.com/leg100/otf/internal"
)

type JobStatus string

const (
	JobPending  JobStatus = "pending"
	JobRunning  JobStatus = "running"
	JobFinished JobStatus = "finished"
	JobErrored  JobStatus = "errored"
)

type Job struct {
	// ID of the run that this job is for.
	RunID string
	// Phase of run that this job is for.
	Phase internal.PhaseType
	// Current status of job.
	Status JobStatus
	// ID of agent that this job is assigned to.
	AgentID string
}

//func newJob(run *otfrun.Run, agentID string) *Job {
//	return &Job{
//		RunID:   run.ID,
//		Phase:   run.Phase(),
//		Status:  JobPending,
//		AgentID: agentID,
//	}
//}
