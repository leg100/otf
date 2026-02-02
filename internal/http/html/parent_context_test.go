package html

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestParentContext(t *testing.T) {
	organizationID1 := resource.NewTfeID(resource.OrganizationKind)
	workspaceID1 := resource.NewTfeID(resource.WorkspaceKind)
	vcsProviderID1 := resource.NewTfeID(resource.VCSProviderKind)
	runID1 := resource.NewTfeID(resource.RunKind)

	pc := ParentContext{
		resolver: &fakeParentContextResolver{
			organizations: map[resource.ID]resource.ID{
				vcsProviderID1: organizationID1,
				runID1:         organizationID1,
			},
			workspaces: map[resource.ID]resource.ID{
				runID1: workspaceID1,
			},
		},
		// Always return the same workspace name, regardless of the ID.
		workspaces: &fakeParentContextWorkspaceClient{
			name: "dev",
		},
	}

	tests := []struct {
		name             string
		path             string
		wantOrganization string
		wantWorkspace    string
	}{
		{
			name:             "organization page",
			path:             fmt.Sprintf("/app/organizations/%v", organizationID1),
			wantOrganization: organizationID1.String(),
		},
		{
			name:             "workspace page",
			path:             fmt.Sprintf("/app/organizations/%v/workspaces/%v", organizationID1, workspaceID1),
			wantOrganization: organizationID1.String(),
			wantWorkspace:    workspaceID1.String(),
		},
		{
			name:             "vcs provider page",
			path:             fmt.Sprintf("/app/vcs-providers/%v", vcsProviderID1),
			wantOrganization: organizationID1.String(),
		},
		{
			name:             "run page",
			path:             fmt.Sprintf("/app/runs/%v", runID1),
			wantOrganization: organizationID1.String(),
			wantWorkspace:    "dev",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("", tt.path, nil)
			r.Header.Set("Content-type", "text/html")
			w := httptest.NewRecorder()

			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				org, _ := OrganizationFromContext(r.Context())
				assert.Equal(t, tt.wantOrganization, org)
				ws, _ := WorkspaceFromContext(r.Context())
				assert.Equal(t, tt.wantWorkspace, ws)
			})

			pc.middleware(h).ServeHTTP(w, r)
		})
	}
}

type fakeParentContextResolver struct {
	organizations map[resource.ID]resource.ID
	workspaces    map[resource.ID]resource.ID
}

func (f *fakeParentContextResolver) GetParentOrganizationID(ctx context.Context, id resource.ID) (resource.ID, error) {
	id, ok := f.organizations[id]
	if !ok {
		return nil, internal.ErrResourceNotFound
	}
	return id, nil
}

func (f *fakeParentContextResolver) GetParentWorkspaceID(ctx context.Context, id resource.ID) (resource.ID, error) {
	id, ok := f.workspaces[id]
	if !ok {
		return nil, internal.ErrResourceNotFound
	}
	return id, nil
}

type fakeParentContextWorkspaceClient struct {
	name string
}

func (f *fakeParentContextWorkspaceClient) GetName(workspaceID resource.TfeID) (string, error) {
	return f.name, nil
}
