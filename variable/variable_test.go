package variable

import (
	"os"
	"path"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateVariable(t *testing.T) {
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
			opts: UpdateVariableOptions{Key: otf.String("teddy")},
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
			opts: UpdateVariableOptions{Value: otf.String("baz")},
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
			opts: UpdateVariableOptions{Sensitive: otf.Bool(true)},
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
			opts: UpdateVariableOptions{HCL: otf.Bool(true)},
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
			opts: UpdateVariableOptions{Sensitive: otf.Bool(false)},
			before: Variable{
				Key:       "foo",
				Value:     "bar",
				Category:  CategoryTerraform,
				Sensitive: true,
			},
			err: true,
		},
		{
			name: "sensitive to non-sensitive",
			opts: UpdateVariableOptions{Sensitive: otf.Bool(false)},
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
			got := tt.before
			err := got.Update(tt.opts)
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

	v1 := NewTestVariable(t, "ws-123", CreateVariableOptions{
		Key:      otf.String("foo"),
		Value:    otf.String("bar"),
		Category: VariableCategoryPtr(CategoryTerraform),
	})
	v2 := NewTestVariable(t, "ws-123", CreateVariableOptions{
		Key: otf.String("images"),
		Value: otf.String(`{
    us-east-1 = "image-1234"
    us-west-2 = "image-4567"
}
`),
		Category: VariableCategoryPtr(CategoryTerraform),
		HCL:      otf.Bool(true),
	})

	err := WriteTerraformVars(dir, []*Variable{v1, v2})
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
