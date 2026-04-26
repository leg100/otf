package path

import (
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
)

var (
	wsID  = resource.MustHardcodeTfeID(resource.WorkspaceKind, "abc123")
	orgID = resource.MustHardcodeTfeID(resource.OrganizationKind, "def456")
	runID = resource.MustHardcodeTfeID(resource.RunKind, "ghi789")
	iaID  = resource.MustHardcodeTfeID(resource.IngressAttributesKind, "jkl012")
)

func TestResource(t *testing.T) {
	tests := []struct {
		name   string
		action resource.Action
		id     resource.ID
		want   string
	}{
		{
			name:   "get omits action",
			action: resource.Get,
			id:     wsID,
			want:   "/app/workspaces/ws-abc123",
		},
		{
			name:   "edit includes action",
			action: resource.Edit,
			id:     wsID,
			want:   "/app/workspaces/ws-abc123/edit",
		},
		{
			name:   "update includes action",
			action: resource.Update,
			id:     wsID,
			want:   "/app/workspaces/ws-abc123/update",
		},
		{
			name:   "delete includes action",
			action: resource.Delete,
			id:     wsID,
			want:   "/app/workspaces/ws-abc123/delete",
		},
		{
			name:   "run resource",
			action: resource.Get,
			id:     runID,
			want:   "/app/runs/run-ghi789",
		},
		{
			name:   "kind with trailing 's' doesn't get another 's' added",
			action: resource.Get,
			id:     iaID,
			want:   "/app/ingress-attributes/ia-jkl012",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Resource(tt.action, tt.id))
		})
	}
}

func TestResources(t *testing.T) {
	tests := []struct {
		name     string
		action   resource.Action
		kind     resource.Kind
		parentID resource.ID
		want     string
	}{
		{
			name:     "list omits action",
			action:   resource.List,
			kind:     resource.WorkspaceKind,
			parentID: orgID,
			want:     "/app/organizations/org-def456/workspaces",
		},
		{
			name:     "new includes action",
			action:   resource.New,
			kind:     resource.WorkspaceKind,
			parentID: orgID,
			want:     "/app/organizations/org-def456/workspaces/new",
		},
		{
			name:     "create includes action",
			action:   resource.Create,
			kind:     resource.WorkspaceKind,
			parentID: orgID,
			want:     "/app/organizations/org-def456/workspaces/create",
		},
		{
			name:     "nil parent omits parent segment",
			action:   resource.List,
			kind:     resource.WorkspaceKind,
			parentID: nil,
			want:     "/app/workspaces",
		},
		{
			name:     "nil parent with non-list action",
			action:   resource.Create,
			kind:     resource.RunKind,
			parentID: nil,
			want:     "/app/runs/create",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Resources(tt.action, tt.kind, tt.parentID))
		})
	}
}

func TestConvenienceFunctions(t *testing.T) {
	assert.Equal(t, "/app/workspaces/ws-abc123", Get(wsID))
	assert.Equal(t, "/app/workspaces/ws-abc123/edit", Edit(wsID))
	assert.Equal(t, "/app/workspaces/ws-abc123/update", Update(wsID))
	assert.Equal(t, "/app/workspaces/ws-abc123/delete", Delete(wsID))
	assert.Equal(t, "/app/organizations/org-def456/workspaces", List(resource.WorkspaceKind, orgID))
	assert.Equal(t, "/app/organizations/org-def456/workspaces/new", New(resource.WorkspaceKind, orgID))
	assert.Equal(t, "/app/organizations/org-def456/workspaces/create", Create(resource.WorkspaceKind, orgID))
}
