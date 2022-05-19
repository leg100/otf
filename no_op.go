package otf

import "context"

// noOp implements a Job
var _ Job = (*noOp)(nil)

// noOp is a job that does nothing
type noOp struct{}

func (_ *noOp) Do(Environment) error              { return nil }
func (_ *noOp) GetID() string                     { return "no-op" }
func (_ *noOp) GetService(Application) JobService { return &noOpService{} }

type noOpService struct{}

func (_ *noOpService) Claim(_ context.Context, _ string, _ JobClaimOptions) (Job, error) {
	return &noOp{}, nil
}

func (_ *noOpService) Finish(_ context.Context, _ string, _ JobFinishOptions) (Job, error) {
	return &noOp{}, nil
}

func (_ *noOpService) PutChunk(_ context.Context, _ string, _ Chunk) error {
	return nil
}
