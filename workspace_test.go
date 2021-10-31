package otf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRun_UpdateStatus tests that UpdateStatus correctly updates the status of
// the run's plan and apply (there is little point to testing the status of the
// run itself because there is no conditional logic to this assignment).
func TestWorkspaceSpecifier_Valid(t *testing.T) {
	tests := []struct {
		name    string
		spec    WorkspaceSpecifier
		invalid bool
	}{
		{
			name: "valid id",
			spec: WorkspaceSpecifier{ID: String("ws-123")},
		},
		{
			name: "valid organization and workspace name",
			spec: WorkspaceSpecifier{OrganizationName: String("org-123"), Name: String("default")},
		},
		{
			name:    "nothing specified",
			spec:    WorkspaceSpecifier{},
			invalid: true,
		},
		{
			name:    "empty id",
			spec:    WorkspaceSpecifier{ID: String("")},
			invalid: true,
		},
		{
			name:    "organization name but no workspace name",
			spec:    WorkspaceSpecifier{OrganizationName: String("org-123")},
			invalid: true,
		},
		{
			name:    "workspace name but no organization name",
			spec:    WorkspaceSpecifier{Name: String("default")},
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
