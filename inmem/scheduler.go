package inmem

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	tfe "github.com/leg100/go-tfe"
	"github.com/leg100/otf"
)

const (
	// SchedulerSubscriptionID is the ID the scheduler uses to identify itself
	// when subscribing to the events service
	SchedulerSubscriptionID = "scheduler"
)

type RunLister interface {
	List(otf.RunListOptions) (*otf.RunList, error)
}

// Scheduler manages workspaces' run queues in memory. It subscribes to events
// and updates the queues accordingly.
type Scheduler struct {
	// Queues is a mapping of workspace ID to workspace queue of runs
	Queues map[string]otf.Queue
	// RunService retrieves and updates runs
	otf.RunService
	// EventService permits scheduler to subscribe to a stream of events
	otf.EventService
	// Logger for logging various events
	logr.Logger
}

// NewScheduler constructs scheduler queues and populates them with existing
// runs.
func NewScheduler(ws otf.WorkspaceService, rs otf.RunService, es otf.EventService, logger logr.Logger) (*Scheduler, error) {
	queues := make(map[string]otf.Queue)

	// Get workspaces
	workspaces, err := ws.List(otf.WorkspaceListOptions{})
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

		queues[ws.ID] = &otf.WorkspaceQueue{PlanEnqueuer: rs, Active: active, Pending: pending}
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

func (s *Scheduler) handleEvent(ev otf.Event) {
	switch obj := ev.Payload.(type) {
	case *otf.Workspace:
		switch ev.Type {
		case otf.WorkspaceCreated:
			s.Queues[obj.ID] = &otf.WorkspaceQueue{PlanEnqueuer: s.RunService}
		case otf.WorkspaceDeleted:
			delete(s.Queues, obj.ID)
		}
	case *otf.Run:
		queue := s.Queues[obj.Workspace.ID]

		switch ev.Type {
		case otf.RunCreated:
			if err := queue.Add(obj); err != nil {
				s.Error(err, "unable to enqueue run", "run", obj.ID)
			}
		case otf.RunCompleted:
			if err := queue.Remove(obj); err != nil {
				s.Error(err, "unable to dequeue run", "run", obj.ID)
			}
		}
	}
}

// getActiveRun retrieves the active (non-speculative) run for the workspace
func getActiveRun(workspaceID string, rl RunLister) (*otf.Run, error) {
	opts := otf.RunListOptions{
		WorkspaceID: &workspaceID,
		Statuses:    otf.ActiveRunStatuses,
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
func filterNonSpeculativeRuns(runs []*otf.Run) (nonSpeculative []*otf.Run) {
	for _, r := range runs {
		if !r.IsSpeculative() {
			nonSpeculative = append(nonSpeculative, r)
		}
	}
	return nonSpeculative
}

// getPendingRuns retrieves pending runs for a workspace
func getPendingRuns(workspaceID string, rl RunLister) ([]*otf.Run, error) {
	opts := otf.RunListOptions{
		WorkspaceID: &workspaceID,
		Statuses:    []tfe.RunStatus{tfe.RunPending},
	}
	pending, err := rl.List(opts)
	if err != nil {
		return nil, err
	}

	return pending.Items, nil
}
