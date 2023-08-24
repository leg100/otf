package variable

import (
	"os"
	"path"
	"testing"

	"github.com/leg100/otf/internal"
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
