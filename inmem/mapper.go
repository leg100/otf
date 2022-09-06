package inmem

import (
	"context"

	"github.com/leg100/otf"
)

// Mapper is an in-memory implementation of a mapper.
//
// A mapper maintains mappings between various resource identifiers, which are
// used by upstream layers to make decisions and efficiently lookup resources.
//
// For instance, the authorization layer needs to decide whether to permit
// access and cannot do so based on a single identifier (e.g. a run id) but
// needs to know which organization and workspace id it relates to.
//
// Whereas the persistence layer, with access to mappings, need only lookup
// resources based on the most appropriate identifier for which it maintains an
// index, rather having to support lookups using a multitude of identifiers.
type Mapper struct {
	workspaces *workspaceMapper
	runs       *runMapper
}

// NewMapper constructs and populates the mapper
func NewMapper() *Mapper {
	return &Mapper{
		workspaces: newWorkspaceMapper(),
		runs:       newRunMapper(),
	}
}

// Populate populates the mapper with identifiers
func (m *Mapper) Populate(ws otf.WorkspaceService, rs otf.RunService) error {
	if err := m.workspaces.populate(ws); err != nil {
		return err
	}
	if err := m.runs.populate(rs); err != nil {
		return err
	}
	return nil
}

// MapRun adds a mapping for a run
func (m *Mapper) MapRun(run *otf.Run) {
	m.runs.add(run)
}

// UnmapRun removes the mapping for a run
func (m *Mapper) UnmapRun(run *otf.Run) {
	m.runs.remove(run)
}

// MapWorkspace adds a mapping for a workspace
func (m *Mapper) MapWorkspace(ws *otf.Workspace) {
	m.workspaces.add(ws)
}

// RemapWorkspace updates a mapping for a workspace
func (m *Mapper) RemapWorkspace(oldName string, ws *otf.Workspace) {
	m.workspaces.update(oldName, ws)
}

// UnmapWorkspace removes a mapping for the workspace
func (m *Mapper) UnmapWorkspace(ws *otf.Workspace) {
	m.workspaces.remove(ws)
}

// LookupWorkspaceID looks up the ID corresponding to the given spec. If the
// spec already contains an ID then that is returned, otherwise the mapper looks
// up the ID corresponding to the given organization and workspace name. If the
// spec is invalid, then an empty string is returned.
func (m *Mapper) LookupWorkspaceID(spec otf.WorkspaceSpec) string {
	if spec.ID != nil {
		return *spec.ID
	} else if spec.OrganizationName != nil && spec.Name != nil {
		return m.workspaces.lookupID(*spec.OrganizationName, *spec.Name)
	} else {
		return ""
	}
}

// CanAccessRun determines if the caller is permitted to access the run
func (m *Mapper) CanAccessRun(ctx context.Context, runID string) bool {
	orgName := m.lookupRunOrganization(runID)
	return otf.CanAccess(ctx, &orgName)
}

// CanAccessWorkspace determines if the caller is permitted to access the
// workspace specified by the spec.
func (m *Mapper) CanAccessWorkspace(ctx context.Context, spec otf.WorkspaceSpec) bool {
	orgName, ok := m.workspaces.lookupOrganizationBySpec(spec)
	if !ok {
		return false
	}
	return otf.CanAccess(ctx, &orgName)
}

// lookupRunOrganization returns a run's organization name given a run ID
func (m *Mapper) lookupRunOrganization(runID string) string {
	workspaceID := m.runs.lookupWorkspaceID(runID)
	return m.workspaces.lookupOrganizationByID(workspaceID)
}
