package otf

import (
	"context"

	"github.com/go-logr/logr"
	"gopkg.in/cenkalti/backoff.v1"
)

// SchedulerLockID is shared by one or more schedulers and is used to guarantee
// that only one scheduler will run at any time.
const SchedulerLockID int64 = 5577006791947779410

// ExclusiveScheduler runs a scheduler, ensuring it is the *only* scheduler
// running.
func ExclusiveScheduler(ctx context.Context, logger logr.Logger, app LockableApplication) {
	op := func() error {
		for {
			err := app.WithLock(ctx, SchedulerLockID, func(app Application) error {
				logger.Info("running scheduler")
				err := NewScheduler(logger, app).Start(ctx)
				return err
			})
			select {
			case <-ctx.Done():
				logger.Info("scheduler gracefully shutdown")
				return nil
			default:
				return err
			}
		}
	}
	backoff.RetryNotify(op, backoff.NewExponentialBackOff(), nil)
}
