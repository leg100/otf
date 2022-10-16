package inmem

import (
	"context"
	"fmt"
	"sync"

	"github.com/leg100/otf"
)

type runMapper struct {
	mu sync.Mutex
	// map run id to workspace id
	idWorkspaceMap map[string]string
}

func newRunMapper() *runMapper {
	return &runMapper{
		idWorkspaceMap: make(map[string]string),
	}
}

func (m *runMapper) populate(ctx context.Context, svc otf.RunService) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	opts := otf.RunListOptions{ListOptions: otf.ListOptions{PageSize: 100}}
	var allocated bool
	for {
		listing, err := svc.ListRuns(ctx, opts)
		if err != nil {
			return fmt.Errorf("populating workspace mapper: %w", err)
		}
		if !allocated {
			m.idWorkspaceMap = make(map[string]string, listing.TotalCount())
			allocated = true
		}
		for _, run := range listing.Items {
			m.idWorkspaceMap[run.ID()] = run.WorkspaceID()
		}
		if listing.NextPage() == nil {
			break
		}
		opts.PageNumber = *listing.NextPage()
	}
	return nil
}

func (m *runMapper) add(run *otf.Run) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.idWorkspaceMap[run.ID()] = run.WorkspaceID()
}

func (m *runMapper) remove(run *otf.Run) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.idWorkspaceMap, run.ID())
}

func (m *runMapper) LookupWorkspaceID(runID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.idWorkspaceMap[runID]
}
