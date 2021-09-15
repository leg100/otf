package agent

import "github.com/leg100/ots"

var _ Job = (*RunJob)(nil)

// Job provides an abstraction of a run
type Job interface {
	GetID() string
	GetStatus() string
	Do(*ots.Environment) error
}

// RunJob wraps a Run to help it implement Job
type RunJob struct {
	*ots.Run
}

func (r *RunJob) GetID() string {
	return r.ID
}

func (r *RunJob) GetStatus() string {
	return string(r.Status)
}
