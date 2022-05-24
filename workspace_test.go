package otf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRun_UpdateStatus tests that updateStatus correctly updates the status of
// the run's plan and apply (there is little point to testing the status of the
// run itself because there is no conditional logic to this assignment).
func TestWorkspaceSpec_Valid(t *testing.T) {
	tests := []struct {
		name    string
		spec    WorkspaceSpec
		invalid bool
	}{
		{
			name: "valid id",
			spec: WorkspaceSpec{ID: String("ws-123")},
		},
		{
			name: "valid organization and workspace name",
			spec: WorkspaceSpec{OrganizationName: String("org-123"), Name: String("default")},
		},
		{
			name:    "nothing specified",
			spec:    WorkspaceSpec{},
			invalid: true,
		},
		{
			name:    "empty id",
			spec:    WorkspaceSpec{ID: String("")},
			invalid: true,
		},
		{
			name:    "organization name but no workspace name",
			spec:    WorkspaceSpec{OrganizationName: String("org-123")},
			invalid: true,
		},
		{
			name:    "workspace name but no organization name",
			spec:    WorkspaceSpec{Name: String("default")},
			invalid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Valid()
			if tt.invalid {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
