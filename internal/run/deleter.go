package run

import (
	"context"
	"time"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
)

type (
	// Deleter deletes old runs that are older than a user-specified age.
	Deleter struct {
		logr.Logger

		OverrideCheckInterval time.Duration
		AgeThreshold          time.Duration
		Runs                  deleterRunClient
	}

	deleterRunClient interface {
		List(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error)
		Delete(ctx context.Context, runID resource.TfeID) error
	}
)

// Start the deleter daemon.
func (e *Deleter) Start(ctx context.Context) error {
	interval := defaultCheckInterval
	if e.OverrideCheckInterval != 0 {
		interval = e.OverrideCheckInterval
	}

	if err := e.deleteRuns(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := e.deleteRuns(ctx); err != nil {
				return err
			}
		}
	}
}

func (e *Deleter) deleteRuns(ctx context.Context) error {
	// Refuse to delete runs if age threshold is set to 0.
	if e.AgeThreshold == 0 {
		return nil
	}
	// Retrieve all runs older than the given age.
	cutoff := time.Now().Add(-e.AgeThreshold)
	runs, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*Run], error) {
		return e.Runs.List(ctx, ListOptions{
			PageOptions:     opts,
			BeforeCreatedAt: &cutoff,
		})
	})
	if err != nil {
		e.Error(err, "retrieving old runs for deletion")
		return err
	}
	for _, run := range runs {
		if err := e.Runs.Delete(ctx, run.ID); err != nil {
			e.Error(err, "deleting old run")
			return err
		}
		e.Info("deleted old run", "run", run)
	}
	return nil
}
