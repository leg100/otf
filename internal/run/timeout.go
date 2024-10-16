package run

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"golang.org/x/exp/maps"
)

// TimeoutLockID is a unique ID guaranteeing only one timeout daemon on a cluster is running at any time.
const TimeoutLockID int64 = 179366396344335598

type (
	// Timeout is a daemon that "times out" runs if one of the phases -
	// planning, applying - exceeds a timeout. This can happen for a number of
	// reasons, for example a terraform plan or apply is stuck talking to an
	// unresponsive API, or if OTF itself has terminated ungracefully and left
	// runs in a planning or applying state.
	Timeout struct {
		logr.Logger

		PlanningTimeout time.Duration
		ApplyingTimeout time.Duration
		Runs            timeoutRunClient
	}

	timeoutRunClient interface {
		List(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error)
		FinishPhase(ctx context.Context, runID string, phase internal.PhaseType, opts PhaseFinishOptions) (*Run, error)
	}
)

// Start the timeout daemon.
func (e *Timeout) Start(ctx context.Context) error {
	// Check every minute for timed out runs.
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			e.check(ctx)
		}
	}
}

func (e *Timeout) check(ctx context.Context) {
	// Statuses that are checked for timeout
	statuses := map[Status]struct {
		// run phase corresponding to phase
		phase   internal.PhaseType
		timeout time.Duration
	}{
		RunPlanning: {
			phase:   internal.PlanPhase,
			timeout: e.PlanningTimeout,
		},
		RunApplying: {
			phase:   internal.ApplyPhase,
			timeout: e.ApplyingTimeout,
		},
	}
	// Retrieve all runs with the given statuses
	runs, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*Run], error) {
		return e.Runs.List(ctx, ListOptions{
			Statuses:    maps.Keys(statuses),
			PageOptions: opts,
		})
	})
	if err != nil {
		e.Error(err, "checking run status timeouts")
		return
	}
	for _, run := range runs {
		s, ok := statuses[run.Status]
		if !ok {
			// Should never happen.
			continue
		}
		// For each run retrieve the timestamp for when it started
		// the status
		started, err := run.StatusTimestamp(run.Status)
		if err != nil {
			// should never happen
			e.Error(err, fmt.Sprintf("checking %s timeout", run.Status))
			continue
		}
		// Check whether the timeout has been exceeded
		if time.Since(started) > s.timeout {
			// Timeout exceeded...
			//
			// Inform the user via log message,
			e.Error(fmt.Errorf("checking %s timeout", run.Status), "timeout", s.timeout, "started", started)
			// And terminate the corresponding phase and the run, forcing it into
			// an errored state.
			//
			// TODO: bubble up to the UI/API the reason for entering the errored
			// state.
			// NOTE: returned error is ignored because FinishPhase
			// logs errors
			_, _ = e.Runs.FinishPhase(ctx, run.ID, s.phase, PhaseFinishOptions{
				Errored: true,
			})
		}
	}
}
