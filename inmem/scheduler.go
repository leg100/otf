package inmem

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	tfe "github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

const (
	// SchedulerSubscriptionID is the ID the scheduler uses to identify itself
	// when subscribing to the events service
	SchedulerSubscriptionID = "scheduler"
)

type RunLister interface {
	List(ots.RunListOptions) (*ots.RunList, error)
}

// Scheduler manages workspaces' run queues in memory. It subscribes to events
// and updates the queues accordingly.
type Scheduler struct {
	// Queues is a mapping of org ID to workspace ID to workspace queue of runs
	Queues map[string]map[string]ots.Queue
	// RunService retrieves and updates runs
	ots.RunService
	// EventService permits scheduler to subscribe to a stream of events
	ots.EventService
	// Logger for logging various events
	logr.Logger
}

// NewScheduler constructs scheduler queues and populates them with existing
// runs.
func NewScheduler(os ots.OrganizationService, ws ots.WorkspaceService, rs ots.RunService, es ots.EventService, logger logr.Logger) (*Scheduler, error) {
	queues := make(map[string]map[string]ots.Queue)

	// Get organizations
	organizations, err := os.List(tfe.OrganizationListOptions{})
	if err != nil {
		return nil, err
	}

	for _, org := range organizations.Items {
		queues[org.ID] = make(map[string]ots.Queue)

		// Get workspaces
		workspaces, err := ws.List(org.ID, tfe.WorkspaceListOptions{})
		if err != nil {
			return nil, err
		}

		for _, ws := range workspaces.Items {
			// Get runs
			active, err := getActiveRun(ws.ID, rs)
			if err != nil {
				return nil, err
			}

			pending, err := getPendingRuns(ws.ID, rs)
			if err != nil {
				return nil, err
			}

			queues[org.ID][ws.ID] = &ots.WorkspaceQueue{RunStatusUpdater: rs, Active: active, Pending: pending}
		}
	}

	s := &Scheduler{
		Queues:       queues,
		RunService:   rs,
		EventService: es,
		Logger:       logger,
	}

	return s, nil
}

// Start starts the scheduler event loop
func (s *Scheduler) Start(ctx context.Context) {
	sub := s.Subscribe(SchedulerSubscriptionID)
	defer sub.Close()

	for {
		select {
		case event, ok := <-sub.C():
			// If sub closed then exit.
			if !ok {
				return
			}
			s.handleEvent(event)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) handleEvent(ev ots.Event) {
	switch obj := ev.Payload.(type) {
	case *ots.Organization:
		switch ev.Type {
		case ots.OrganizationCreated:
			s.Queues[obj.ID] = make(map[string]ots.Queue)
		case ots.OrganizationDeleted:
			delete(s.Queues, obj.ID)
		}
	case *ots.Workspace:
		switch ev.Type {
		case ots.WorkspaceCreated:
			s.Queues[obj.Organization.ID][obj.ID] = &ots.WorkspaceQueue{RunStatusUpdater: s.RunService}
		case ots.WorkspaceDeleted:
			delete(s.Queues[obj.Organization.ID], obj.ID)
		}
	case *ots.Run:
		queue := s.Queues[obj.Workspace.Organization.ID][obj.Workspace.ID]

		switch ev.Type {
		case ots.RunCreated:
			if err := queue.Add(obj); err != nil {
				s.Error(err, "unable to enqueue run", "run", obj.ID)
			}
		case ots.RunCompleted:
			if err := queue.Remove(obj); err != nil {
				s.Error(err, "unable to dequeue run", "run", obj.ID)
			}
		}
	}
}

// getActiveRun retrieves the active (non-speculative) run for the workspace
func getActiveRun(workspaceID string, rl RunLister) (*ots.Run, error) {
	opts := ots.RunListOptions{
		WorkspaceID: &workspaceID,
		Statuses:    ots.ActiveRunStatuses,
	}
	active, err := rl.List(opts)
	if err != nil {
		return nil, err
	}

	nonSpeculative := filterNonSpeculativeRuns(active.Items)
	switch len(nonSpeculative) {
	case 0:
		return nil, nil
	case 1:
		return nonSpeculative[0], nil
	default:
		return nil, fmt.Errorf("more than one active non-speculative run found")
	}

}

// filterNonSpeculativeRuns filters out speculative runs
func filterNonSpeculativeRuns(runs []*ots.Run) (nonSpeculative []*ots.Run) {
	for _, r := range runs {
		if !r.IsSpeculative() {
			nonSpeculative = append(nonSpeculative, r)
		}
	}
	return nonSpeculative
}

// getPendingRuns retrieves pending runs for a workspace
func getPendingRuns(workspaceID string, rl RunLister) ([]*ots.Run, error) {
	opts := ots.RunListOptions{
		WorkspaceID: &workspaceID,
		Statuses:    []tfe.RunStatus{tfe.RunPending},
	}
	pending, err := rl.List(opts)
	if err != nil {
		return nil, err
	}

	return pending.Items, nil
}
