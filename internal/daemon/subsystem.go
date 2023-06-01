package daemon

import (
	"context"
	"errors"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"golang.org/x/sync/errgroup"
	"gopkg.in/cenkalti/backoff.v1"
)

type (
	// Subsystem is an automonous system subordinate to the main daemon (otfd).
	Subsystem struct {
		// Name of subsystem
		Name string
		// System is the underlying system to be invoked and supervised.
		System Startable
		// Backoff and restart subsystem in the event of an error
		BackoffRestart bool
		// Exclusive: permit only one instance of this subsystem on an OTF
		// cluster
		Exclusive bool
		// DB for obtaining cluster-wide lock. Must be non-nil if Exclusive is
		// true.
		internal.DB
		// Cluster-unique lock ID. Must be non-nil if Exclusive is true.
		LockID *int64
		logr.Logger
	}
	// Startable is a blocking process that is started at least once, and upon error,
	// may need re-starting.
	Startable interface {
		Start(ctx context.Context) error
	}
)

func (s *Subsystem) Start(ctx context.Context, g *errgroup.Group) error {
	if s.Exclusive {
		if s.LockID == nil {
			return errors.New("exclusive subsystem must have non-nil lock ID")
		}
		if s.DB == nil {
			return errors.New("exclusive subsystem must have non-nil database")
		}
	}

	// Confer all privileges to subsystem and identify subsystem in service
	// endpoint calls.
	ctx = internal.AddSubjectToContext(ctx, &internal.Superuser{Username: s.Name})

	op := func() error {
		if s.Exclusive {
			// block on getting an exclusive lock
			return s.WaitAndLock(ctx, *s.LockID, func() error {
				return s.System.Start(ctx)
			})
		} else {
			return s.System.Start(ctx)
		}
	}
	if s.BackoffRestart {
		// Backoff and retry whenever operation returns an error. If context is
		// cancelled then it'll stop retrying and return the context error.
		policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
		g.Go(func() error {
			return backoff.RetryNotify(op, policy, func(err error, next time.Duration) {
				// re-open semaphore
				s.Error(err, "restarting "+s.Name)
			})
		})
	} else {
		g.Go(op)
	}
	s.V(1).Info("started subsystem", "name", s.Name)
	return nil
}
