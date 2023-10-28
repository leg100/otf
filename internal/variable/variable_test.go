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

func TestVariable_Update(t *testing.T) {
	tests := []struct {
		name   string
		opts   UpdateVariableOptions
		before Variable
		after  Variable
		err    bool // want error
	}{
		{
			name: "no change",
			opts: UpdateVariableOptions{},
			before: Variable{
				Key:      "foo",
				Value:    "bar",
				Category: CategoryTerraform,
			},
			after: Variable{
				Key:      "foo",
				Value:    "bar",
				Category: CategoryTerraform,
			},
		},
		{
			name: "key",
			opts: UpdateVariableOptions{Key: internal.String("teddy")},
			before: Variable{
				Key:      "foo",
				Value:    "bar",
				Category: CategoryTerraform,
			},
			after: Variable{
				Key:      "teddy",
				Value:    "bar",
				Category: CategoryTerraform,
			},
		},
		{
			name: "value",
			opts: UpdateVariableOptions{Value: internal.String("baz")},
			before: Variable{
				Key:      "foo",
				Value:    "bar",
				Category: CategoryTerraform,
			},
			after: Variable{
				Key:      "foo",
				Value:    "baz",
				Category: CategoryTerraform,
			},
		},
		{
			name: "non-sensitive to sensitive",
			opts: UpdateVariableOptions{Sensitive: internal.Bool(true)},
			before: Variable{
				Key:      "foo",
				Value:    "bar",
				Category: CategoryTerraform,
			},
			after: Variable{
				Key:       "foo",
				Value:     "bar",
				Category:  CategoryTerraform,
				Sensitive: true,
			},
		},
		{
			name: "non-hcl to hcl",
			opts: UpdateVariableOptions{HCL: internal.Bool(true)},
			before: Variable{
				Key:      "foo",
				Value:    "bar",
				Category: CategoryTerraform,
			},
			after: Variable{
				Key:      "foo",
				Value:    "bar",
				Category: CategoryTerraform,
				HCL:      true,
			},
		},
		{
			name: "sensitive to non-sensitive",
			opts: UpdateVariableOptions{Sensitive: internal.Bool(false)},
			before: Variable{
				Key:       "foo",
				Value:     "bar",
				Category:  CategoryTerraform,
				Sensitive: true,
			},
			err: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.opts.generateVersion = func() string { return "" }
			got := tt.before
			err := got.update(nil, tt.opts)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.after, got)
			}
		})
	}
}

func TestWriteTerraformVariables(t *testing.T) {
	dir := t.TempDir()

	v1, err := newVariable(nil, CreateVariableOptions{
		Key:      internal.String("foo"),
		Value:    internal.String("bar"),
		Category: VariableCategoryPtr(CategoryTerraform),
	})
	require.NoError(t, err)

	v2, err := newVariable(nil, CreateVariableOptions{
		Key: internal.String("images"),
		Value: internal.String(`{
    us-east-1 = "image-1234"
    us-west-2 = "image-4567"
}
`),
		Category: VariableCategoryPtr(CategoryTerraform),
		HCL:      internal.Bool(true),
	})
	require.NoError(t, err)

	v3, err := newVariable(nil, CreateVariableOptions{
		Key:      internal.String("multiline-foo"),
		Value:    internal.String("foo\nbar\nbaz"),
		Category: VariableCategoryPtr(CategoryTerraform),
	})
	require.NoError(t, err)

	v4, err := newVariable(nil, CreateVariableOptions{
		Key:      internal.String("multiline-foo-with-delimiter"),
		Value:    internal.String("EOTfoo\nbar\nbaz"),
		Category: VariableCategoryPtr(CategoryTerraform),
	})
	require.NoError(t, err)

	err = WriteTerraformVars(dir, []*Variable{v1, v2, v3, v4})
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
multiline-foo = <<EOT
foo
bar
baz
EOT
multiline-foo-with-delimiter = <<EOTT
EOTfoo
bar
baz
EOTT
`
	assert.Equal(t, want, string(got))
}

func Test_mergeVariables(t *testing.T) {
	tests := []struct {
		name               string
		sets               []*VariableSet
		workspaceVariables []*Variable
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
			workspaceVariables: []*Variable{
				{
					Key:      "workspace",
					Value:    "true",
					Category: CategoryTerraform,
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
		// a variable set can be set both to global and also specify workspaces,
		// in which case the workspaces should be ignored and not considered as
		// part of determining precedence.
		{
			name: "ignore workspaces in global sets",
			sets: []*VariableSet{
				{
					// even though this has lexical precedence, it is global and
					// thus have lower precedence than the workspace-scoped set
					// below.
					Name:       "a - global with workspaces",
					Global:     true,
					Workspaces: []string{"ws-123"},
					Variables: []*Variable{
						{
							Key:      "foo",
							Value:    "global",
							Category: CategoryTerraform,
						},
					},
				},
				{
					Name:       "b - workspace-scoped",
					Workspaces: []string{"ws-123"},
					Variables: []*Variable{
						{
							Key:      "foo",
							Value:    "workspace-scoped",
							Category: CategoryTerraform,
						},
					},
				},
			},
			want: []*Variable{
				{
					Key:      "foo",
					Value:    "workspace-scoped",
					Category: CategoryTerraform,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeVariables(tt.sets, tt.workspaceVariables, &tt.run)
			assert.Equal(t, len(tt.want), len(got))
			for _, w := range tt.want {
				assert.Contains(t, got, w)
			}
		})
	}
}
