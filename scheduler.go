package ots

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	tfe "github.com/leg100/go-tfe"
)

const (
	// SchedulerSubscriptionID is the ID the scheduler uses to identify itself
	// when subscribing to the events service
	SchedulerSubscriptionID = "scheduler"
)

type RunLister interface {
	List(string, RunListOptions) (*RunList, error)
}

// Scheduler schedules workspace's incomplete runs.
type Scheduler struct {
	// Queues is a mapping of org ID to workspace ID to workspace queue of runs
	Queues map[string]map[string]*WorkspaceQueue
	// RunService retrieves and updates runs
	RunService
	// EventService permits scheduler to subscribe to a stream of events
	EventService
	// Logger for logging various events
	log logr.Logger
}

// NewScheduler constructs a Scheduler
func NewScheduler(os OrganizationService, rs RunService, ws WorkspaceService, es EventService, logger logr.Logger) (*Scheduler, error) {
	queues := make(map[string]map[string]*WorkspaceQueue)

	// Get organizations
	organizations, err := os.List(tfe.OrganizationListOptions{})
	if err != nil {
		return nil, err
	}

	for _, org := range organizations.Items {
		queues[org.ID] = make(map[string]*WorkspaceQueue)

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

			queues[org.ID][ws.ID] = NewWorkspaceQueue(rs, logger, ws.ID, WithActive(active), WithPending(pending))
		}
	}

	s := &Scheduler{
		Queues:       queues,
		RunService:   rs,
		EventService: es,
		log:          logger,
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

func (s *Scheduler) handleEvent(ev Event) {
	switch obj := ev.Payload.(type) {
	case *Organization:
		switch ev.Type {
		case OrganizationCreated:
			s.Queues[obj.ID] = make(map[string]*WorkspaceQueue)
		case OrganizationDeleted:
			delete(s.Queues, obj.ID)
		}
	case *Workspace:
		switch ev.Type {
		case WorkspaceCreated:
			s.Queues[obj.Organization.ID][obj.ID] = &WorkspaceQueue{}
		case WorkspaceDeleted:
			delete(s.Queues[obj.Organization.ID], obj.ID)
		}
	case *Run:
		switch ev.Type {
		case RunCreated:
			s.Queues[obj.Workspace.Organization.ID][obj.Workspace.ID].addRun(obj)
		case RunCompleted:
			s.Queues[obj.Workspace.Organization.ID][obj.Workspace.ID].removeRun(obj)
		}
	}
}

// getActiveRun retrieves the active (non-speculative) run for the workspace
func getActiveRun(workspaceID string, rl RunLister) (*Run, error) {
	opts := RunListOptions{
		Statuses: ActiveRunStatuses,
	}
	active, err := rl.List(workspaceID, opts)
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

func filterNonSpeculativeRuns(runs []*Run) (nonSpeculative []*Run) {
	for _, r := range runs {
		if !r.IsSpeculative() {
			nonSpeculative = append(nonSpeculative, r)
		}
	}
	return nonSpeculative
}

// getPendingRuns retrieves pending runs for a workspace
func getPendingRuns(workspaceID string, rl RunLister) ([]*Run, error) {
	opts := RunListOptions{
		Statuses: []tfe.RunStatus{tfe.RunPending},
	}
	pending, err := rl.List(workspaceID, opts)
	if err != nil {
		return nil, err
	}

	return pending.Items, nil
}
