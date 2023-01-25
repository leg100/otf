package scheduler

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"gopkg.in/cenkalti/backoff.v1"
)

// schedulerLockID guarantees only one scheduler on a cluster is running at any
// time.
const schedulerLockID int64 = 5577006791947779410

// ExclusiveScheduler runs a scheduler, ensuring it is the *only* scheduler
// running.
func ExclusiveScheduler(ctx context.Context, logger logr.Logger, app otf.LockableApplication) error {
	op := func() error {
		for {
			err := app.WithLock(ctx, schedulerLockID, func(app otf.Application) error {
				return newScheduler(logger, app).start(ctx)
			})
			select {
			case <-ctx.Done():
				return nil
			default:
				return err
			}
		}
	}
	return backoff.RetryNotify(op, backoff.NewExponentialBackOff(), nil)
}
