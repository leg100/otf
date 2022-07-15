package app

import (
	"context"

	"github.com/leg100/otf"
)

type Mapper struct {
	// map workspace id to organization name
	workspaceIDs map[string]string
	// map workspace name to workspace id
	workspaceNames map[string]string
	// map run id to workspace id
	runIDs map[string]string
}

func (m *Mapper) AddWorkspace(ws *otf.Workspace) {
	m.workspaceIDs[ws.ID()] = ws.OrganizationName()
	m.workspaceNames[ws.Name()] = ws.ID()
}

func (m *Mapper) RemoveWorkspace(ws *otf.Workspace) {
	delete(m.workspaceIDs, ws.ID())
	delete(m.workspaceNames, ws.Name())
}

func (m *Mapper) LookupRunOrganization(runID string) (string, bool) {
	workspaceID, ok := m.runIDs[runID]
	if !ok {
		return "", false
	}
	orgName, ok := m.workspaceIDs[workspaceID]
	return orgName, ok
}

// helper that takes a run ID and looks up its organization name to determine if
// the caller has permission for the organization
func (m *Mapper) CanAccessRun(ctx context.Context, runID string) bool {
	orgName, ok := m.LookupRunOrganization(runID)
	if !ok {
		return false
	}
	return otf.CanAccess(ctx, orgName)
}
