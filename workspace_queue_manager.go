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

// workspaceQueueManager manages workspace queues of runs
type workspaceQueueManager struct {
	// RunService retrieves and updates runs
	RunService
	// EventService permits scheduler to subscribe to a stream of events
	EventService
	// Logger for logging messages
	logr.Logger
	// run queue for each workspace
	queues map[string]workspaceQueue
	// context to terminate manager
	ctx context.Context
}

// NewWorkspaceQueueManager constructs and populates workspace queues with runs
// before starting any eligible runs.
func NewWorkspaceQueueManager(ctx context.Context, app Application, logger logr.Logger) (*workspaceQueueManager, error) {
	s := &workspaceQueueManager{
		RunService:   app.RunService(),
		EventService: app.EventService(),
		Logger:       logger.WithValues("component", "workspace_queue_manager"),
		ctx:          ctx,
		queues:       make(map[string]workspaceQueue),
	}
	if err := s.seed(); err != nil {
		return nil, err
	}
	if err := s.startRuns(); err != nil {
		return nil, err
	}
	return s, nil
}

// Start the scheduler event loop and manage queues in response to events
func (s *workspaceQueueManager) Start() error {
	sub, err := s.Subscribe(SchedulerSubscriptionID)
	if err != nil {
		return err
	}
	defer sub.Close()
	for {
		select {
		case <-s.ctx.Done():
			return nil
		case event, ok := <-sub.C():
			if !ok {
				return nil
			}
			switch obj := event.Payload.(type) {
			case *Workspace:
				switch event.Type {
				case EventWorkspaceDeleted:
					// garbage collect queue
					delete(s.queues, obj.ID())
				}
				// ignore EventWorkspaceCreated because the mgr creates
				// workspace queue on-demand when a run event comes in.
			case *Run:
				s.refresh(obj)
			}
		}
	}
}

// seed populates workspace queues
func (s *workspaceQueueManager) seed() error {
	runs, err := s.List(s.ctx, RunListOptions{Statuses: IncompleteRun})
	if err != nil {
		return err
	}
	for _, run := range runs.Items {
		if run.Speculative() {
			// speculative runs are never queued
			continue
		}
		s.update(run.Workspace.ID(), func(q workspaceQueue) (workspaceQueue, error) {
			q = append(q, run)
			return q, nil
		})
	}
	return nil
}

// startRuns inspects the front of each workspace queue and starts eligible runs
func (s *workspaceQueueManager) startRuns() error {
	for workspaceID := range s.queues {
		err := s.update(workspaceID, func(q workspaceQueue) (workspaceQueue, error) {
			return q.startRun(s.ctx, s.RunService)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// refresh takes an updated run and refreshes its workspace queue, starting an
// eligible run if there is one.
func (s *workspaceQueueManager) refresh(run *Run) error {
	if run.Speculative() {
		// speculative runs are never queued
		return nil
	}
	// update workspace queue and start eligible run if there is one
	err := s.update(run.Workspace.ID(), func(q workspaceQueue) (workspaceQueue, error) {
		return q.update(run).startRun(s.ctx, s.RunService)
	})
	return err
}

// update map of queues in-place, updating the workspace queue using the
// supplied fn. The workspace queue is created if it doesn't exist.
func (s *workspaceQueueManager) update(workspaceID string, fn func(workspaceQueue) (workspaceQueue, error)) error {
	q, ok := s.queues[workspaceID]
	if !ok {
		q = workspaceQueue{}
	}
	q, err := fn(q)
	if err != nil {
		return err
	}
	s.queues[workspaceID] = q
	return nil
}
