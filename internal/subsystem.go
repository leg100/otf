package internal

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"gopkg.in/cenkalti/backoff.v1"
)

type (
	Subsystem interface {
		Initialize(context.Context) error
	}
	SubsystemWrapper struct {
		// Name of subsystem
		Name string
		// Backoff and retry initialization the subsystem in the event of an error
		BackoffRetry bool
		// Exclusive: permit only one instance of this subsystem on an OTF
		// cluster
		Exclusive bool
		// DB for obtaining cluster-wide lock. Must be non-nil if Exclusive is
		// true.
		DB
		// Cluster-unique lock ID. Must be non-nil if Exclusive is true.
		LockID *int64
		// Subsystem that this wrapper wraps
		Subsystem
		logr.Logger
	}
)

func (s *SubsystemWrapper) Start(ctx context.Context) error {
	op := func() error {
		if s.Exclusive {
			// block on getting an exclusive lock
			return s.WaitAndLock(ctx, *s.LockID, func() error {
				return s.Initialize(ctx)
			})
		} else {
			return s.Initialize(ctx)
		}
	}
	if s.BackoffRetry {
		// Backoff and retry whenever operation returns an error. If context is
		// cancelled then it'll stop retrying and return the context error.
		policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
		return backoff.RetryNotify(op, policy, func(err error, next time.Duration) {
			s.Error(err, "restarting "+s.Name)
		})
	} else {
		return op()
	}
}