package variable

import (
	"os"
	"path"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteTerraformVariables(t *testing.T) {
	dir := t.TempDir()

	v1 := newTestVariable(t, "ws-123", CreateVariableOptions{
		Key:      internal.String("foo"),
		Value:    internal.String("bar"),
		Category: VariableCategoryPtr(CategoryTerraform),
	})
	v2 := newTestVariable(t, "ws-123", CreateVariableOptions{
		Key: internal.String("images"),
		Value: internal.String(`{
    us-east-1 = "image-1234"
    us-west-2 = "image-4567"
}
`),
		Category: VariableCategoryPtr(CategoryTerraform),
		HCL:      internal.Bool(true),
	})

	err := WriteTerraformVars(dir, []*Variable{v1.Variable, v2.Variable})
	require.NoError(t, err)

	tfvars := path.Join(dir, "terraform.tfvars")
	got, err := os.ReadFile(tfvars)
	require.NoError(t, err)

	want := `
foo = "bar"
images = {
    us-east-1 = "image-1234"
    us-west-2 = "image-4567"
}
`
	assert.Equal(t, want, string(got))
}

func Test_mergeVariables(t *testing.T) {
	tests := []struct {
		name               string
		sets               []*VariableSet
		workspaceVariables []*WorkspaceVariable
		run                run.Run
		want               []*Variable
	}{
		{
			name: "default",
			sets: []*VariableSet{
				{
					Name:   "global",
					Global: true,
					Variables: []*Variable{
						{
							Key:      "global",
							Value:    "true",
							Category: CategoryTerraform,
						},
					},
				},
				{
					Name:       "workspace-scoped",
					Workspaces: []string{"ws-123"},
					Variables: []*Variable{
						{
							Key:      "workspace-scoped",
							Value:    "true",
							Category: CategoryTerraform,
						},
					},
				},
			},
			workspaceVariables: []*WorkspaceVariable{
				{
					Variable: &Variable{
						Key:      "workspace",
						Value:    "true",
						Category: CategoryTerraform,
					},
				},
			},
			run: run.Run{WorkspaceID: "ws-123", Variables: []run.Variable{{Key: "run", Value: "true"}}},
			want: []*Variable{
				{
					Key:      "global",
					Value:    "true",
					Category: CategoryTerraform,
				},
				{
					Key:      "workspace-scoped",
					Value:    "true",
					Category: CategoryTerraform,
				},
				{
					Key:      "workspace",
					Value:    "true",
					Category: CategoryTerraform,
				},
				{
					Key:      "run",
					Value:    "true",
					Category: CategoryTerraform,
					HCL:      true,
				},
			},
		},
		{
			name: "workspace-scoped set lexical precedence",
			sets: []*VariableSet{
				{
					Name:       "set_A",
					Workspaces: []string{"ws-123"},
					Variables: []*Variable{
						{
							Key:      "foo",
							Value:    "set_a",
							Category: CategoryTerraform,
						},
					},
				},
				{
					Name:       "set_B",
					Workspaces: []string{"ws-123"},
					Variables: []*Variable{
						{
							Key:      "foo",
							Value:    "set_b",
							Category: CategoryTerraform,
						},
					},
				},
			},
			run: run.Run{WorkspaceID: "ws-123"},
			want: []*Variable{
				{
					Key:      "foo",
					Value:    "set_a",
					Category: CategoryTerraform,
				},
			},
		},
		{
			name: "skip sets scoped to a different workspace",
			sets: []*VariableSet{
				{
					Name:       "set_A",
					Workspaces: []string{"ws-123"},
					Variables: []*Variable{
						{
							Key:      "foo",
							Value:    "set_a",
							Category: CategoryTerraform,
						},
					},
				},
				{
					Name:       "set_B",
					Workspaces: []string{"ws-456"},
					Variables: []*Variable{
						{
							Key:      "foo",
							Value:    "set_b",
							Category: CategoryTerraform,
						},
					},
				},
			},
			run: run.Run{WorkspaceID: "ws-456"},
			want: []*Variable{
				{
					Key:      "foo",
					Value:    "set_b",
					Category: CategoryTerraform,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeVariables(&tt.run, tt.workspaceVariables, tt.sets)
			assert.Equal(t, tt.want, got)
		})
	}
}
