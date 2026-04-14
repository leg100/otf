package daemon

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/logr"
	"golang.org/x/sync/errgroup"
)

type (
	// Subsystem is an automonous system subordinate to the main daemon (otfd).
	Subsystem struct {
		Logger logr.Logger
		// Name of subsystem
		Name string
		// System is the underlying system to be invoked and supervised.
		System Startable
		// Exclusive if true ensures only one instance of this subsystem is
		// started in a multi-node otfd cluster.
		Exclusive bool
	}
	// Startable is a blocking process that is started at least once, and upon error,
	// may need re-starting.
	Startable interface {
		Start(ctx context.Context) error
	}

	//subsystemDB interface {
	//	WaitAndLock(ctx context.Context, id int64, fn func(context.Context) error) error
	//}
)

func (s *Subsystem) Start(ctx context.Context, g *errgroup.Group) error {
	// Confer all privileges to subsystem and identify subsystem in service
	// endpoint calls.
	ctx = authz.AddSubjectToContext(ctx, &authz.Superuser{Username: s.Name})

	op := func() error {
		s.Logger.V(1).Info("started subsystem", "name", s.Name)
		err := s.System.Start(ctx)
		if ctx.Err() != nil {
			// don't return an error if subsystem was terminated via a
			// canceled context.
			s.Logger.V(1).Info("gracefully shutdown subsystem", "name", s.Name)
			return nil
		}
		if err != nil {
			return fmt.Errorf("subsystem prematurely exited: %w", err)
		}
		return fmt.Errorf("subsystem prematurely exited without an error")
	}
	// Backoff and retry whenever operation returns an error. If context is
	// cancelled then it'll stop retrying and return the context error.
	infiniteRetry := backoff.WithMaxElapsedTime(0)
	policy := backoff.WithContext(backoff.NewExponentialBackOff(infiniteRetry), ctx)
	g.Go(func() error {
		return backoff.RetryNotify(op, policy, func(err error, next time.Duration) {
			s.Logger.Error(err, "restarting subsystem", "name", s.Name, "backoff", next)
		})
	})
	return nil
}
