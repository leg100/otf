package daemon

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal/authz"
	"golang.org/x/sync/errgroup"
)

type (
	// Subsystem is an automonous system subordinate to the main daemon (otfd).
	Subsystem struct {
		// Name of subsystem
		Name string
		// System is the underlying system to be invoked and supervised.
		System Startable
		// Exclusive: permit only one instance of this subsystem on an OTF
		// cluster
		Exclusive bool
		// DB for obtaining cluster-wide lock. Must be non-nil if Exclusive is
		// true.
		DB subsystemDB
		// Cluster-unique lock ID. Must be non-nil if Exclusive is true.
		LockID *int64
		logr.Logger
	}
	// Startable is a blocking process that is started at least once, and upon error,
	// may need re-starting.
	Startable interface {
		Start(ctx context.Context) error
	}

	subsystemDB interface {
		WaitAndLock(ctx context.Context, id int64, fn func(context.Context) error) error
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
	ctx = authz.AddSubjectToContext(ctx, &authz.Superuser{Username: s.Name})

	op := func() (err error) {
		if s.Exclusive {
			// block on getting an exclusive lock
			err = s.DB.WaitAndLock(ctx, *s.LockID, func(ctx context.Context) error {
				return s.System.Start(ctx)
			})
		} else {
			err = s.System.Start(ctx)
		}
		if ctx.Err() != nil {
			// don't return an error if subsystem was terminated via a
			// canceled context.
			s.V(1).Info("gracefully shutdown subsystem", "name", s.Name)
			return nil
		}
		return err
	}
	// Backoff and retry whenever operation returns an error. If context is
	// cancelled then it'll stop retrying and return the context error.
	infiniteRetry := backoff.WithMaxElapsedTime(0)
	policy := backoff.WithContext(backoff.NewExponentialBackOff(infiniteRetry), ctx)
	g.Go(func() error {
		return backoff.RetryNotify(op, policy, func(err error, next time.Duration) {
			s.Error(err, "restarting subsystem", "name", s.Name, "backoff", next)
		})
	})
	s.V(1).Info("started subsystem", "name", s.Name)
	return nil
}
