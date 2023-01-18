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

func TestBuildSandboxArgs(t *testing.T) {
	t.Run("without plugin cache", func(t *testing.T) {
		env := Environment{
			Terraform: &fakeTerraform{"/bins"},
			path:      "/root",
		}
		want := []string{
			"--ro-bind", "/bins/terraform", "/bin/terraform",
			"--bind", "/root", "/workspace",
			"--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf",
			"--ro-bind", "/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/certs/ca-certificates.crt",
			"--chdir", "/workspace",
			"--proc", "/proc",
			"terraform", "apply",
			"-input=false", "-no-color",
		}
		assert.Equal(t, want, env.buildSandboxArgs([]string{"-input=false", "-no-color"}))
	})

	t.Run("with plugin cache", func(t *testing.T) {
		env := Environment{
			Terraform: &fakeTerraform{"/bins"},
			path:      "/root",
			Config: Config{
				PluginCache: true,
			},
		}
		want := []string{
			"--ro-bind", "/bins/terraform", "/bin/terraform",
			"--bind", "/root", "/workspace",
			"--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf",
			"--ro-bind", "/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/certs/ca-certificates.crt",
			"--chdir", "/workspace",
			"--proc", "/proc",
			"--ro-bind", PluginCacheDir, PluginCacheDir,
			"terraform", "apply",
			"-input=false", "-no-color",
		}
		assert.Equal(t, want, env.buildSandboxArgs([]string{"-input=false", "-no-color"}))
	})
}
