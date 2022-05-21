package otf

import (
	"context"

	"github.com/go-logr/logr"
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
func NewWorkspaceQueueManager(ctx context.Context, app Application, logger logr.Logger) (*workspaceQueueManager, error) {
	s := &workspaceQueueManager{
		RunService:   app.RunService(),
		EventService: app.EventService(),
		Logger:       logger.WithValues("component", "workspace_queue_manager"),
		ctx:          ctx,
	}
	if err := s.seed(); err != nil {
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
				case EventWorkspaceCreated:
					// create workspace queue
					s.queues[obj.ID] = workspaceQueue{}
				case EventWorkspaceDeleted:
					// delete workspace queue
					delete(s.queues, obj.ID)
				}
			case *Run:
				if obj.IsSpeculative() {
					// speculative runs are never queued
					continue
				}
				// update workspace queue and start run
				s.update(obj.Workspace.ID, func(q workspaceQueue) workspaceQueue {
					q, err = q.update(obj).startRun(s.ctx, s.RunService)
					if err != nil {
						s.Error(err, "starting run")
					}
					return q
				})
			}
		}
	}
}

// seed populates workspace queues and starts runs at the front of queues.
func (s *workspaceQueueManager) seed() error {
	runs, err := s.List(s.ctx, RunListOptions{Statuses: IncompleteRunStatuses})
	if err != nil {
		return err
	}
	for _, run := range runs.Items {
		if run.IsSpeculative() {
			// speculative runs are never queued
			continue
		}
		s.update(run.Workspace.ID, func(q workspaceQueue) workspaceQueue {
			return append(q, run)
		})
	}
	for workspaceID := range s.queues {
		s.update(workspaceID, func(q workspaceQueue) workspaceQueue {
			q, err := q.startRun(s.ctx, s.RunService)
			if err != nil {
				s.Error(err, "starting run")
			}
			return q
		})
	}
	return nil
}

// update map of queues in-place, updating the queue for the specified workspace
// using the supplied fn. The workspace queue is created if it doesn't exist.
func (s *workspaceQueueManager) update(workspaceID string, fn func(workspaceQueue) workspaceQueue) {
	q, ok := s.queues[workspaceID]
	if !ok {
		q = workspaceQueue{}
	}
	s.queues[workspaceID] = fn(q)
}
