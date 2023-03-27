package variable

import (
	"os"
	"path"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
