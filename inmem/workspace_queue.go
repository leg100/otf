package inmem

import (
	"fmt"

	"github.com/leg100/otf"
)

// WorkspaceQueue is an in-memory workspace queue. Not concurrency-safe.
type WorkspaceQueue struct {
	// Ordered map implementation
	mapping map[string]int
	order   []*otf.Run
}

func NewWorkspaceQueue() *WorkspaceQueue {
	return &WorkspaceQueue{
		order:   make([]*otf.Run, 0),
		mapping: make(map[string]int),
	}
}

// Update queue with a run.
func (q *WorkspaceQueue) Update(run *otf.Run) {
	if run.Speculative() {
		return
	}
	if pos := q.position(run); pos >= 0 {
		if run.Done() {
			q.remove(pos)
		} else {
			// update run in-place
			q.order[pos] = run
		}
	} else {
		if !run.Done() {
			q.add(run)
		}
	}
}

// Get queue of runs
func (q *WorkspaceQueue) Get() []*otf.Run {
	return q.order
}

func (q *WorkspaceQueue) add(run *otf.Run) {
	q.order = append(q.order, run)
	q.mapping[run.ID()] = len(q.order) - 1
}

func (q *WorkspaceQueue) remove(pos int) {
	// remove run from queue
	q.order = append(q.order[:pos], q.order[pos+1:]...)
	// we need to rebuild the map of positions as they may all now be incorrect
	newMapping := make(map[string]int, len(q.order))
	for pos, run := range q.order {
		newMapping[run.ID()] = pos
	}
	q.mapping = newMapping
}

func (q *WorkspaceQueue) position(run *otf.Run) int {
	pos, ok := q.mapping[run.ID()]
	if ok {
		return pos
	}
	return -1
}

// WorkspaceQueueManager manages workspace queues
type WorkspaceQueueManager struct {
	// mapping of workspace ID to queue
	queues map[string]*WorkspaceQueue
}

func NewWorkspaceQueueManager() *WorkspaceQueueManager {
	return &WorkspaceQueueManager{queues: make(map[string]*WorkspaceQueue)}
}

// Populate creates queues for each workspace.
func (m *WorkspaceQueueManager) Populate(svc otf.WorkspaceService) error {
	opts := otf.WorkspaceListOptions{}
	var allocated bool
	for {
		listing, err := svc.ListWorkspaces(otf.ContextWithAppUser(), opts)
		if err != nil {
			return fmt.Errorf("populating workspace mapper: %w", err)
		}
		if !allocated {
			m.queues = make(map[string]*WorkspaceQueue, listing.TotalCount())
			allocated = true
		}
		for _, ws := range listing.Items {
			m.Create(ws.ID())
		}
		if listing.NextPage() == nil {
			break
		}
		opts.PageNumber = *listing.NextPage()
	}
	return nil
}

// Create a new queue for the workspace with the given ID.
func (m *WorkspaceQueueManager) Create(workspaceID string) {
	m.queues[workspaceID] = NewWorkspaceQueue()
}

func (m *WorkspaceQueueManager) Update(workspaceID string, run *otf.Run) error {
	q, ok := m.queues[workspaceID]
	if !ok {
		return fmt.Errorf("workspace queue for %s not found", workspaceID)
	}
	q.Update(run)
	return nil
}

func (m *WorkspaceQueueManager) Get(workspaceID string) ([]*otf.Run, error) {
	q, ok := m.queues[workspaceID]
	if !ok {
		return nil, fmt.Errorf("workspace queue for %s not found", workspaceID)
	}
	return q.Get(), nil
}

func (m *WorkspaceQueueManager) Delete(workspaceID string) {
	delete(m.queues, workspaceID)
}
