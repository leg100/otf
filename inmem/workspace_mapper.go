package inmem

import (
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

// newWorkspaceMapper populates workspace mapper
func (m *workspaceMapper) populate(svc otf.WorkspaceService) (*workspaceMapper, error) {
	opts := otf.WorkspaceListOptions{}
	var allocated bool
	for {
		listing, err := svc.List(otf.ContextWithAppUser(), opts)
		if err != nil {
			return nil, fmt.Errorf("populating workspace mapper: %w", err)
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
	return m, nil
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

func (m *workspaceMapper) update(oldName string, ws *otf.Workspace) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nameIDMap[ws.QualifiedName()] = ws.ID()
	delete(m.nameIDMap, otf.WorkspaceQualifiedName{
		Name:         oldName,
		Organization: ws.OrganizationName(),
	})
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
