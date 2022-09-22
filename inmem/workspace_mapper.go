package inmem

import (
	"context"
	"fmt"
	"sync"

	"github.com/leg100/otf"
)

type workspaceMapper struct {
	mu sync.Mutex
	// map workspace id to organization name
	idOrgMap map[string]string
	// map qualified workspace name to workspace id
	nameIDMap map[otf.WorkspaceQualifiedName]string
}

func newWorkspaceMapper() *workspaceMapper {
	return &workspaceMapper{
		idOrgMap:  make(map[string]string),
		nameIDMap: make(map[otf.WorkspaceQualifiedName]string),
	}
}

func (m *workspaceMapper) populate(ctx context.Context, svc otf.WorkspaceService) error {
	opts := otf.WorkspaceListOptions{}
	var allocated bool
	for {
		listing, err := svc.ListWorkspaces(ctx, opts)
		if err != nil {
			return fmt.Errorf("populating workspace mapper: %w", err)
		}
		if !allocated {
			m.idOrgMap = make(map[string]string, listing.TotalCount())
			m.nameIDMap = make(map[otf.WorkspaceQualifiedName]string, listing.TotalCount())
			allocated = true
		}
		for _, ws := range listing.Items {
			m.addWithoutLock(ws)
		}
		if listing.NextPage() == nil {
			break
		}
		opts.PageNumber = *listing.NextPage()
	}
	return nil
}

func (m *workspaceMapper) add(ws *otf.Workspace) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.addWithoutLock(ws)
}

func (m *workspaceMapper) addWithoutLock(ws *otf.Workspace) {
	m.idOrgMap[ws.ID()] = ws.OrganizationName()
	m.nameIDMap[ws.QualifiedName()] = ws.ID()
}

// update the mapping for a workspace that has been renamed
func (m *workspaceMapper) update(ws *otf.Workspace) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// we don't have the old name to hand, so we have to enumerate every entry
	// and look for a workspace with a matching name.
	for qualified, id := range m.nameIDMap {
		if ws.ID() == id {
			// remove old entry
			delete(m.nameIDMap, qualified)
			// add new entry
			m.nameIDMap[ws.QualifiedName()] = ws.ID()
			return
		}
	}
}

// LookupWorkspaceID looks up the ID of the workspace given its name and
// organization name.
func (m *workspaceMapper) lookupID(org, name string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.nameIDMap[otf.WorkspaceQualifiedName{
		Name:         name,
		Organization: org,
	}]
}

func (m *workspaceMapper) remove(ws *otf.Workspace) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.idOrgMap, ws.ID())
	delete(m.nameIDMap, ws.QualifiedName())
}

func (m *workspaceMapper) lookupOrganizationByID(workspaceID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.idOrgMap[workspaceID]
}

func (m *workspaceMapper) lookupOrganizationBySpec(spec otf.WorkspaceSpec) (string, bool) {
	if spec.OrganizationName != nil {
		return *spec.OrganizationName, true
	} else if spec.ID != nil {
		m.mu.Lock()
		defer m.mu.Unlock()

		return m.idOrgMap[*spec.ID], true
	} else {
		return "", false
	}
}
