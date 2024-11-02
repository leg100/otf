package runner

import (
	"context"
	"time"

	"github.com/leg100/otf/internal/authz"
)

var (
	pingTimeout            = 30 * time.Second
	defaultManagerInterval = 10 * time.Second
)

// ManagerLockID guarantees only one manager on a cluster is running at any
// time.
const ManagerLockID int64 = 5577006791947779413

// manager manages the state of runners.
//
// Only one manager should be running on an OTF cluster at any one time.
type manager struct {
	// service for retrieving runners and updating their state.
	client managerClient
	// frequency with which the manager will check runners.
	interval time.Duration
	// manager identifies itself as a subject when making service calls
	authz.Subject
}

type managerClient interface {
	listRunners(ctx context.Context) ([]*RunnerMeta, error)
	updateStatus(ctx context.Context, runnerID string, status RunnerStatus) error
	deleteRunner(ctx context.Context, runnerID string) error
}

func newManager(s *Service) *manager {
	return &manager{
		client:   s,
		interval: defaultManagerInterval,
	}
}

func (m *manager) String() string { return "runner-manager" }

// Start the manager. Every interval the status of runners is checked,
// updating their status as necessary.
//
// Should be invoked in a go routine.
func (m *manager) Start(ctx context.Context) error {
	ctx = authz.AddSubjectToContext(ctx, m)

	updateAll := func() error {
		runners, err := m.client.listRunners(ctx)
		if err != nil {
			return err
		}
		for _, runner := range runners {
			if err := m.update(ctx, runner); err != nil {
				return err
			}
		}
		return nil
	}
	// run at startup and then every x seconds
	if err := updateAll(); err != nil {
		return err
	}
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := updateAll(); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (m *manager) update(ctx context.Context, runner *RunnerMeta) error {
	switch runner.Status {
	case RunnerIdle, RunnerBusy:
		// update runner status to unknown if the runner has failed to ping within
		// the timeout.
		if time.Since(runner.LastPingAt) > pingTimeout {
			return m.client.updateStatus(ctx, runner.ID, RunnerUnknown)
		}
	case RunnerUnknown:
		// update runner status from unknown to errored if a further period of 5
		// minutes has elapsed.
		if time.Since(runner.LastStatusAt) > 5*time.Minute {
			// update runner status to errored.
			return m.client.updateStatus(ctx, runner.ID, RunnerErrored)
		}
	case RunnerErrored, RunnerExited:
		// purge runner from database once a further 1 hour has elapsed for
		// runners in a terminal state.
		if time.Since(runner.LastStatusAt) > time.Hour {
			return m.client.deleteRunner(ctx, runner.ID)
		}
	}
	return nil
}
