package otf

import "context"

type RunCreator interface {
	CreateRun(ctx context.Context, run *Run) (*Run, error)
}
