package resource

import (
	"context"
	"time"

	"github.com/leg100/otf/internal/logr"
)

// By default check resources every minute
var deleterDefaultCheckInterval = time.Minute

type (
	deleteableResource interface {
		GetID() TfeID
	}

	// Deleter deletes resources that are older than a user-specified age.
	Deleter[R deleteableResource] struct {
		logr.Logger

		OverrideCheckInterval time.Duration
		AgeThreshold          time.Duration
		Client                deleterClient[R]
	}

	deleterClient[R any] interface {
		ListOlderThan(ctx context.Context, age time.Time) ([]R, error)
		Delete(ctx context.Context, id TfeID) error
	}
)

// Start the deleter daemon.
func (e *Deleter[R]) Start(ctx context.Context) error {
	interval := deleterDefaultCheckInterval
	if e.OverrideCheckInterval != 0 {
		interval = e.OverrideCheckInterval
	}

	if err := e.deleteResources(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := e.deleteResources(ctx); err != nil {
				return err
			}
		}
	}
}

func (e *Deleter[R]) deleteResources(ctx context.Context) error {
	// Refuse to delete resources if age threshold is set to 0.
	if e.AgeThreshold == 0 {
		return nil
	}
	// Retrieve all resources older than the given age.
	cutoff := time.Now().Add(-e.AgeThreshold)
	resources, err := e.Client.ListOlderThan(ctx, cutoff)
	if err != nil {
		e.Error(err, "retrieving old resources for deletion")
		return err
	}
	for _, res := range resources {
		if err := e.Client.Delete(ctx, res.GetID()); err != nil {
			e.Error(err, "deleting old resource")
			return err
		}
		e.Info("deleted old resource", "resource", res)
	}
	return nil
}
