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
func (m *Mapper) Populate(ws otf.WorkspaceService, rs otf.RunService) (err error) {
	workspaces, err := m.workspaces.populate(ws)
	if err != nil {
		return err
	}
	runs, err := m.runs.populate(rs)
	if err != nil {
		return err
	}
	m.workspaces = workspaces
	m.runs = runs
	return nil
}

func (m *Mapper) AddRun(run *otf.Run) {
	m.runs.add(run)
}

func (m *Mapper) RemoveRun(run *otf.Run) {
	m.runs.remove(run)
}

func (m *Mapper) AddWorkspace(ws *otf.Workspace) {
	m.workspaces.add(ws)
}

func (m *Mapper) UpdateWorkspace(oldName string, ws *otf.Workspace) {
	m.workspaces.update(oldName, ws)
}

func (m *Mapper) RemoveWorkspace(ws *otf.Workspace) {
	m.workspaces.remove(ws)
}

func (m *Mapper) LookupWorkspaceID(org, name string) string {
	return m.workspaces.lookupID(org, name)
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
