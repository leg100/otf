package runner

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/resource"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

type ExecutorKind string

var _ pflag.Value = (*ExecutorKind)(nil)

func (*ExecutorKind) Type() string         { return "executor" }
func (e *ExecutorKind) Set(v string) error { return e.set(v) }
func (e *ExecutorKind) String() string     { return string(*e) }

func (e *ExecutorKind) set(v string) error {
	switch ExecutorKind(v) {
	case ForkExecutorKind, KubeExecutorKind:
	default:
		return fmt.Errorf("no executor named %s: must be %s or %s", v, ForkExecutorKind, KubeExecutorKind)
	}
	*e = ExecutorKind(v)
	return nil
}

type executor interface {
	// SpawnOperation spawns an operation to carry out a job.
	SpawnOperation(ctx context.Context, g *errgroup.Group, job *Job, jobToken []byte) error
	// currentJobs returns the number of current jobs.
	currentJobs(ctx context.Context, runnerID resource.TfeID) int
}
