package otf

import "context"

// noOp implements a Job
var _ Job = (*noOp)(nil)

// noOp is a job that does nothing
type noOp struct{}

func (*noOp) Do(Environment) error              { return nil }
func (*noOp) GetID() string                     { return "no-op" }
func (*noOp) GetService(Application) JobService { return &noOpService{} }

type noOpService struct{}

func (*noOpService) Claim(_ context.Context, _ string, _ JobClaimOptions) (Job, error) {
	return &noOp{}, nil
}

func (*noOpService) Finish(_ context.Context, _ string, _ JobFinishOptions) (Job, error) {
	return &noOp{}, nil
}

func (*noOpService) PutChunk(_ context.Context, _ string, _ Chunk) error {
	return nil
}
