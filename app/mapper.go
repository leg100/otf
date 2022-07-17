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

func NewMapper() *Mapper {
	return &Mapper{
		workspaceIDs:   make(map[string]string),
		workspaceNames: make(map[string]string),
		runIDs:         make(map[string]string),
	}
}

func (m *Mapper) AddWorkspace(ws *otf.Workspace) {
	m.workspaceIDs[ws.ID()] = ws.OrganizationName()
	m.workspaceNames[ws.Name()] = ws.ID()
}

func (m *Mapper) UpdateWorkspace(oldName string, ws *otf.Workspace) {
	m.workspaceNames[ws.Name()] = ws.ID()
	delete(m.workspaceNames, oldName)
}

func (m *Mapper) RemoveWorkspace(ws *otf.Workspace) {
	delete(m.workspaceIDs, ws.ID())
	delete(m.workspaceNames, ws.Name())
}

func (m *Mapper) AddRun(run *otf.Run) {
	m.runIDs[run.ID()] = run.WorkspaceID()
}

func (m *Mapper) RemoveRun(run *otf.Run) {
	delete(m.runIDs, run.ID())
}

func (m *Mapper) LookupRunOrganization(runID string) (string, bool) {
	workspaceID, ok := m.runIDs[runID]
	if !ok {
		return "", false
	}
	orgName, ok := m.workspaceIDs[workspaceID]
	return orgName, ok
}

func (m *Mapper) LookupWorkspaceOrganization(spec otf.WorkspaceSpec) (string, bool) {
	if spec.OrganizationName != nil {
		return *spec.OrganizationName, true
	} else if spec.ID != nil {
		return m.workspaceIDs[*spec.ID], true
	} else {
		return "", false
	}
}

// CanAccessRun determines if the caller is permitted to access the run
func (m *Mapper) CanAccessRun(ctx context.Context, runID string) bool {
	orgName, ok := m.LookupRunOrganization(runID)
	if !ok {
		return false
	}
	return otf.CanAccess(ctx, &orgName)
}

// CanAccessWorkspace determines if the caller is permitted to access the
// workspace specified by the spec.
func (m *Mapper) CanAccessWorkspace(ctx context.Context, spec otf.WorkspaceSpec) bool {
	orgName, ok := m.LookupWorkspaceOrganization(spec)
	if !ok {
		return false
	}
	return otf.CanAccess(ctx, &orgName)
}
