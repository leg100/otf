package run

import (
	"context"
	"fmt"
	"time"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"golang.org/x/exp/maps"
)

// By default check timed out runs every minute
var defaultCheckInterval = time.Minute

type (
	// Timeout is a daemon that "times out" runs if one of the phases -
	// planning, applying - exceeds a timeout. This can happen for a number of
	// reasons, for example a terraform plan or apply is stuck talking to an
	// unresponsive API, or if OTF itself has terminated ungracefully and left
	// runs in a planning or applying state.
	Timeout struct {
		logr.Logger

		OverrideCheckInterval time.Duration
		PlanningTimeout       time.Duration
		ApplyingTimeout       time.Duration
		Runs                  timeoutRunClient
	}

	timeoutRunClient interface {
		List(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error)
		Cancel(ctx context.Context, runID resource.TfeID) error
	}
)

// Start the timeout daemon.
func (e *Timeout) Start(ctx context.Context) error {
	// Set the interval between checking for timed out runs. Unless an override
	// interval has been provided, use a default.
	interval := defaultCheckInterval
	if e.OverrideCheckInterval != 0 {
		interval = e.OverrideCheckInterval
	}

	ticker := time.NewTicker(interval)
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
	statuses := map[runstatus.Status]struct {
		// phase corresponding to status
		phase PhaseType
		// each status has a specific timeout
		timeout time.Duration
	}{
		runstatus.Planning: {
			phase:   PlanPhase,
			timeout: e.PlanningTimeout,
		},
		runstatus.Applying: {
			phase:   ApplyPhase,
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
			e.Error(err, "checking run timeout", "run_id", run.ID, "status", run.Status)
			continue
		}
		// Check whether the timeout has been exceeded
		if time.Since(started) > s.timeout {
			// Timeout exceeded...
			//
			// Inform the user via log message,
			e.Error(nil, fmt.Sprintf("%s timeout exceeded", run.Status),
				fmt.Sprintf("%s_timeout", run.Status), s.timeout,
				fmt.Sprintf("started_%s", run.Status), started,
				"run_id", run.ID,
			)
			// Send cancellation signal to terminate terraform process and force
			// run into the canceled state.
			//
			// TODO: bubble up to the UI/API the reason for cancelling the run.
			_ = e.Runs.Cancel(ctx, run.ID)
		}
	}
}
