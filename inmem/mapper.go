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
	otf.Application
	workspaces *workspaceMapper
	runs       *runMapper
}

// NewMapper constructs the mapper
func NewMapper(app otf.Application) *Mapper {
	return &Mapper{
		Application: app,
		workspaces:  newWorkspaceMapper(),
		runs:        newRunMapper(),
	}
}

// Start the mapper, populate entries from the DB, and watch changes, updating
// mappings accordingly.
func (m *Mapper) Start(ctx context.Context) error {
	// Register for events first so we don't miss any.
	sub, err := m.Watch(ctx, otf.WatchOptions{})
	if err != nil {
		return err
	}

	if err := m.workspaces.populate(ctx, m.Application); err != nil {
		return err
	}
	if err := m.runs.populate(ctx, m.Application); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-sub:
			if !ok {
				return nil
			}
			switch obj := event.Payload.(type) {
			case *otf.Workspace:
				switch event.Type {
				case otf.EventWorkspaceCreated:
					m.workspaces.add(obj)
				case otf.EventWorkspaceDeleted:
					m.workspaces.remove(obj)
				case otf.EventWorkspaceRenamed:
					m.workspaces.update(obj)
				}
			case *otf.Run:
				switch event.Type {
				case otf.EventRunCreated:
					m.runs.add(obj)
				case otf.EventRunDeleted:
					m.runs.remove(obj)
				}
			}
		}
	}
}

// Populate populates the mapper with identifiers
func (m *Mapper) Populate(ctx context.Context, ws otf.WorkspaceService, rs otf.RunService) error {
	if err := m.workspaces.populate(ctx, ws); err != nil {
		return err
	}
	if err := m.runs.populate(ctx, rs); err != nil {
		return err
	}
	return nil
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
