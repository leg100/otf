package otf

// noOp implements a Job
var _ Job = (*noOp)(nil)

// noOp is a job that does nothing
type noOp struct{}

func (*noOp) Do(Environment) error { return nil }
func (*noOp) JobID() string        { return "no-op" }
