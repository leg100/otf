package otf

import (
	"context"

	"github.com/go-logr/logr"
	"gopkg.in/cenkalti/backoff.v1"
)

// SchedulerLockID guarantees only one scheduler on a cluster is running at any
// time.
const SchedulerLockID int64 = 5577006791947779410

// ExclusiveScheduler runs a scheduler, ensuring it is the *only* scheduler
// running.
func ExclusiveScheduler(ctx context.Context, logger logr.Logger, app LockableApplication) error {
	op := func() error {
		for {
			err := app.WithLock(ctx, SchedulerLockID, func(app Application) error {
				return NewScheduler(logger, app).Start(ctx)
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
