package otf

import "time"

// noOp implements a Job
var _ Job = (*noOp)(nil)

// noOp is a job that does nothing
type noOp struct{}

func (*noOp) Do(Environment) error                      { return nil }
func (*noOp) JobID() string                             { return "no-op" }
func (*noOp) JobStatus() JobStatus                      { return JobPending }
func (*noOp) JobStatusTimestamps() []JobStatusTimestamp { return []JobStatusTimestamp{} }
func (*noOp) JobStatusTimestamp(JobStatus) (time.Time, error) {
	return time.Now(), nil

}
