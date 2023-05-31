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
	// Subsystem is an automonous system subordinate to the daemon (otfd).
	Subsystem struct {
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
		logr.Logger
		// SubsystemOperation is the underlying the operation that this
		// subsystem invokes.
		SubsystemOperation
	}
	// SubsystemOperation is the operation the subsystem should initialize and
	// supervise, re-initializing if necessary.
	SubsystemOperation interface {
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

	var started chan struct{}
	op := func() error {
		started = make(chan struct{})
		if s.Exclusive {
			// block on getting an exclusive lock
			return s.WaitAndLock(ctx, *s.LockID, func() error {
				return s.SubsystemOperation.Start(ctx, started)
			})
		} else {
			return s.SubsystemOperation.Start(ctx, started)
		}
	}
	if s.BackoffRetry {
		// Backoff and retry whenever operation returns an error. If context is
		// cancelled then it'll stop retrying and return the context error.
		policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
		g.Go(func() error {
			return backoff.RetryNotify(op, policy, func(err error, next time.Duration) {
				s.Error(err, "restarting "+s.Name)
			})
		})
	} else {
		g.Go(op)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second * 10):
		return fmt.Errorf("timed out waiting for %s to start", s.Name)
	case <-started:
		return nil
	}
}
