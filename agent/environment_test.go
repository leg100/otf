package agent

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironment(t *testing.T) {
	env := Environment{
		Logger: logr.Discard(),
		out:    discard{},
	}
	err := env.Execute(&fakeJob{"sleep", []string{"1"}})
	require.NoError(t, err)
}

func TestEnvironment_Cancel(t *testing.T) {
	env := Environment{
		Logger: logr.Discard(),
		out:    discard{},
	}

	wait := make(chan error)
	go func() {
		wait <- env.Execute(&fakeJob{"sleep", []string{"100"}})
	}()
	// give the 'sleep' cmd above time to start
	time.Sleep(100 * time.Millisecond)

	env.Cancel(false)
	err := <-wait
	assert.Error(t, err)
}

func TestWriteTerraformVariables(t *testing.T) {
	dir := t.TempDir()

	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)
	v1 := otf.NewTestVariable(t, ws, otf.CreateVariableOptions{
		Key:      otf.String("foo"),
		Value:    otf.String("bar"),
		Category: otf.VariableCategoryPtr(otf.CategoryTerraform),
	})
	v2 := otf.NewTestVariable(t, ws, otf.CreateVariableOptions{
		Key: otf.String("images"),
		Value: otf.String(`{
    us-east-1 = "image-1234"
    us-west-2 = "image-4567"
}
`),
		Category: otf.VariableCategoryPtr(otf.CategoryTerraform),
		HCL:      otf.Bool(true),
	})

	err := writeTerraformVariables(dir, []*otf.Variable{v1, v2})
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
