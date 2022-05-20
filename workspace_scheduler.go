package otf

import (
	"context"

	"github.com/go-logr/logr"
)

const (
	// SchedulerSubscriptionID is the ID the scheduler uses to identify itself
	// when subscribing to the events service
	SchedulerSubscriptionID = "scheduler"
)

// WorkspaceScheduler starts and queues runs on behalf of workspaces.
type WorkspaceScheduler struct {
	// RunService retrieves and updates runs
	RunService
	// EventService permits scheduler to subscribe to a stream of events
	EventService
	// Logger for logging various events
	logr.Logger
	// run queue for each workspace
	queues map[string][]*Run
}

// NewScheduler seeds workspaces queues and starts any pending speculative runs
func NewWorkspaceScheduler(ctx context.Context, ws WorkspaceService, rs RunService, es EventService, logger logr.Logger) (*WorkspaceScheduler, error) {
	workspaces, err := ws.List(ctx, WorkspaceListOptions{})
	if err != nil {
		return nil, err
	}
	s := &WorkspaceScheduler{
		queues:       make(map[string][]*Run, len(workspaces.Items)),
		RunService:   rs,
		EventService: es,
		Logger:       logger,
	}
	for _, ws := range workspaces.Items {
		s.queues[ws.ID] = []*Run{}
		opts := RunListOptions{
			WorkspaceID: &ws.ID,
			Statuses:    IncompleteRunStatuses,
		}
		incomplete, err := rs.List(ctx, opts)
		if err != nil {
			return nil, err
		}
		for _, run := range incomplete.Items {
			s.refresh(ctx, run)
		}
	}
	return s, nil
}

// Start the scheduler event loop and respond to incoming events
func (s *WorkspaceScheduler) Start(ctx context.Context) error {
	sub, err := s.Subscribe(SchedulerSubscriptionID)
	if err != nil {
		return err
	}
	defer sub.Close()
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-sub.C():
			if !ok {
				return nil
			}
			switch obj := event.Payload.(type) {
			case *Workspace:
				switch event.Type {
				case EventWorkspaceCreated:
					s.queues[obj.ID] = []*Run{}
				case EventWorkspaceDeleted:
					delete(s.queues, obj.ID)
				}
			case *Run:
				s.refresh(ctx, obj)
			}
		}
	}
}

// refresh queues in response to an updated run and start eligible pending run
func (s *WorkspaceScheduler) refresh(ctx context.Context, updated *Run) {
	if updated.IsSpeculative() {
		if updated.Status == RunPending {
			s.RunService.Start(ctx, updated.ID)
		}
		// speculative runs are never enqueued
		return
	}
	wid := updated.Workspace.ID
	if pos := s.position(updated); pos >= 0 {
		if updated.IsDone() {
			s.remove(wid, pos)
		} else {
			s.queues[wid][pos] = updated
		}
	} else {
		// add to queue
		s.queues[wid] = append(s.queues[wid], updated)
	}
	if len(s.queues[wid]) > 0 {
		if s.queues[wid][0].Status == RunPending {
			s.RunService.Start(ctx, s.queues[wid][0].ID)
		}
	}
	return
}

// position retrieves run position in workspace queue
func (s *WorkspaceScheduler) position(run *Run) int {
	wid := run.Workspace.ID
	for i, r := range s.queues[wid] {
		if r.ID == run.ID {
			return i
		}
	}
	return -1
}

// remove run from indexed position in workspace queue
func (s *WorkspaceScheduler) remove(wid string, i int) {
	// remove from queue
	s.queues[wid] = append(s.queues[wid][:i], s.queues[wid][i+1:]...)
}
