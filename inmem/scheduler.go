package inmem

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

const (
	// SchedulerSubscriptionID is the ID the scheduler uses to identify itself
	// when subscribing to the events service
	SchedulerSubscriptionID = "scheduler"
)

type RunLister interface {
	List(context.Context, otf.RunListOptions) (*otf.RunList, error)
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

// NewScheduler constructs workspaces queues and seeds them with existing runs.
func NewScheduler(ws otf.WorkspaceService, rs otf.RunService, es otf.EventService, logger logr.Logger) (*Scheduler, error) {
	workspaces, err := ws.List(context.Background(), otf.WorkspaceListOptions{})
	if err != nil {
		return nil, err
	}
	queues := make(map[string]otf.Queue, len(workspaces.Items))
	for _, ws := range workspaces.Items {
		queues[ws.ID] = &otf.WorkspaceQueue{}
		opts := otf.RunListOptions{
			WorkspaceID: &ws.ID,
			Statuses:    otf.IncompleteRunStatuses,
		}
		incomplete, err := rs.List(context.Background(), opts)
		if err != nil {
			return nil, err
		}
		for _, run := range incomplete.Items {
			if startable := queues[ws.ID].Update(run); startable != nil {
				rs.EnqueuePlan(context.Background(), startable.ID)
			}
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
func (s *Scheduler) Start(ctx context.Context) error {
	sub, err := s.Subscribe(SchedulerSubscriptionID)
	if err != nil {
		return err
	}
	defer sub.Close()
	for {
		select {
		case event, ok := <-sub.C():
			// If sub closed then exit.
			if !ok {
				return nil
			}
			s.handleEvent(event)
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Scheduler) handleEvent(ev otf.Event) {
	switch obj := ev.Payload.(type) {
	case *otf.Workspace:
		switch ev.Type {
		case otf.EventWorkspaceCreated:
			s.Queues[obj.ID] = &otf.WorkspaceQueue{}
		case otf.EventWorkspaceDeleted:
			delete(s.Queues, obj.ID)
		}
	ase *otf.Run:
		queue := s.Queues[obj.Workspace.ID]
		queue.Update(obj)
		if startable := queue.Startable(); startable != nil {
			s.EnqueuePlan(context.Background(), startable.ID)
		}
	}
}
