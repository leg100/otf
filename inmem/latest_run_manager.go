package inmem

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
)

var _ otf.LatestRunService = (*LatestRunManager)(nil)

// LatestRunManager maintains in memory the latest run for each workspace.
type LatestRunManager struct {
	// mapping of workspace ID to ID of latest run
	latest map[string]string
}

func NewLatestRunManager(svc otf.WorkspaceService) (*LatestRunManager, error) {
	m := &LatestRunManager{}

	// Retrieve latest run for each workspace
	opts := otf.WorkspaceListOptions{}
	for {
		listing, err := svc.ListWorkspaces(otf.ContextWithAppUser(), opts)
		if err != nil {
			return nil, fmt.Errorf("retrieving latest runs: %w", err)
		}
		if m.latest == nil {
			m.latest = make(map[string]string, listing.TotalCount())
		}
		for _, ws := range listing.Items {
			if ws.LatestRunID() != nil {
				m.latest[ws.ID()] = *ws.LatestRunID()
			}
		}
		if listing.NextPage() == nil {
			break
		}
		opts.PageNumber = *listing.NextPage()
	}

	return m, nil
}

// SetLatestRun sets the ID of the latest run for a workspace.
func (m *LatestRunManager) SetLatestRun(ctx context.Context, workspaceID string, runID string) error {
	m.latest[workspaceID] = runID
	return nil
}

// GetLatestRun retrieves the ID of the latest run for a workspace.
func (m *LatestRunManager) GetLatestRun(ctx context.Context, workspaceID string) (string, bool) {
	latest, ok := m.latest[workspaceID]
	return latest, ok
}
