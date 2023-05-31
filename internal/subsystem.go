package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
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
		DB
		// Cluster-unique lock ID. Must be non-nil if Exclusive is true.
		LockID *int64
		logr.Logger
	}
	// Startable is a blocking process that closes the started channel once it
	// has successfully started.
	Startable interface {
		Start(ctx context.Context, started chan struct{}) error
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
	ctx = AddSubjectToContext(ctx, &Superuser{Username: s.Name})

	started := make(chan struct{})
	op := func() error {
		if s.Exclusive {
			// block on getting an exclusive lock
			return s.WaitAndLock(ctx, *s.LockID, func() error {
				return s.System.Start(ctx, started)
			})
		} else {
			return s.System.Start(ctx, started)
		}
	}
	if s.BackoffRestart {
		// Backoff and retry whenever operation returns an error. If context is
		// cancelled then it'll stop retrying and return the context error.
		policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
		g.Go(func() error {
			return backoff.RetryNotify(op, policy, func(err error, next time.Duration) {
				// re-open semaphore
				started = make(chan struct{})
				s.Error(err, "restarting "+s.Name)
			})
		})
	} else {
		g.Go(op)
	}
	// Don't wait for an exclusive system to start because it may be waiting for
	// a lock to become free (i.e. another otfd node is already running the system).
	if s.Exclusive {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second * 10):
		return fmt.Errorf("timed out waiting for %s to start", s.Name)
	case <-started:
		s.V(1).Info("started subdaemon", "name", s.Name)
		return nil
	}
}
