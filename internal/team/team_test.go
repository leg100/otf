package team

import (
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTeam_CanAccess_ReadWorkspaces tests that a team granted the
// organization-wide 'read workspaces' permission can view, but not modify,
// workspaces within its organization.
func TestTeam_CanAccess_ReadWorkspaces(t *testing.T) {
	org := organization.NewTestName(t)
	other := organization.NewTestName(t)
	team := &Team{
		ID:             resource.NewTfeID(resource.TeamKind),
		Name:           "viewers",
		Organization:   org,
		ReadWorkspaces: true,
	}
	workspaceID := resource.NewTfeID(resource.WorkspaceKind)

	tests := []struct {
		name   string
		action resource.Action
		kind   resource.Kind
		id     resource.ID
		want   bool
	}{
		// permitted: viewing workspaces and everything belonging to them
		{"list workspaces in organization", resource.List, resource.WorkspaceKind, org, true},
		{"get workspace", resource.Get, resource.WorkspaceKind, workspaceID, true},
		{"list runs", resource.List, resource.RunKind, workspaceID, true},
		{"get run", resource.Get, resource.RunKind, workspaceID, true},
		{"tail logs", resource.Tail, resource.ChunkKind, workspaceID, true},
		{"get plan file", resource.Get, resource.PlanFileKind, workspaceID, true},
		{"get state version", resource.Get, resource.StateVersionKind, workspaceID, true},
		{"download state version", resource.Download, resource.StateVersionKind, workspaceID, true},
		{"get state version output", resource.Get, resource.StateVersionOutputKind, workspaceID, true},

		// denied: anything that starts a run or changes the workspace
		{"create run", resource.Create, resource.RunKind, workspaceID, false},
		{"apply run", resource.Apply, resource.RunKind, workspaceID, false},
		{"cancel run", resource.Cancel, resource.RunKind, workspaceID, false},
		{"create workspace", resource.Create, resource.WorkspaceKind, org, false},
		{"update workspace", resource.Update, resource.WorkspaceKind, workspaceID, false},
		{"delete workspace", resource.Delete, resource.WorkspaceKind, workspaceID, false},
		{"lock workspace", resource.Lock, resource.WorkspaceKind, workspaceID, false},
		{"create variable", resource.Create, resource.VariableKind, workspaceID, false},
		{"update variable", resource.Update, resource.VariableKind, workspaceID, false},
		{"upload logs", resource.Upload, resource.ChunkKind, workspaceID, false},

		// denied: another organization's workspaces
		{"list workspaces in another organization", resource.List, resource.WorkspaceKind, other, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := team.CanAccess(tt.action, tt.kind, authz.Request{ID: tt.id})
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestTeam_ManageWorkspacesImpliesRead tests that granting a team permission to
// manage workspaces implicitly grants permission to read them too, mirroring
// TFC/TFE behaviour.
func TestTeam_ManageWorkspacesImpliesRead(t *testing.T) {
	org := organization.NewTestName(t)

	created, err := NewTeam(org, CreateTeamOptions{
		Name:                      new("managers"),
		OrganizationAccessOptions: OrganizationAccessOptions{ManageWorkspaces: new(true)},
	})
	require.NoError(t, err)
	assert.True(t, created.ReadWorkspaces)

	updated := &Team{Organization: org, Name: "managers"}
	require.NoError(t, updated.Update(UpdateTeamOptions{
		OrganizationAccessOptions: OrganizationAccessOptions{ManageWorkspaces: new(true)},
	}))
	assert.True(t, updated.ReadWorkspaces)
}
