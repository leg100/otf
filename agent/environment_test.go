package agent

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/registry"
	"github.com/leg100/otf/variable"
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

func TestEnvironment_WorkingDir(t *testing.T) {
	tests := []struct {
		name    string
		workdir string
	}{
		{
			"default working dir", "",
		},
		{
			"custom working dir", "subdir",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnvironment(t, otf.WorkingDirectory(tt.workdir))
			assert.Equal(t, tt.workdir, env.relWorkDir)
			assert.DirExists(t, env.rootDir)
			assert.DirExists(t, env.absWorkDir)
			assert.FileExists(t, path.Join(env.absWorkDir, "terraform.tfvars"))
		})
	}
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
	v1 := variable.NewTestVariable(t, ws, otf.CreateVariableOptions{
		Key:      otf.String("foo"),
		Value:    otf.String("bar"),
		Category: variable.VariableCategoryPtr(otf.CategoryTerraform),
	})
	v2 := variable.NewTestVariable(t, ws, otf.CreateVariableOptions{
		Key: otf.String("images"),
		Value: otf.String(`{
    us-east-1 = "image-1234"
    us-west-2 = "image-4567"
}
`),
		Category: variable.VariableCategoryPtr(otf.CategoryTerraform),
		HCL:      otf.Bool(true),
	})

	err := writeTerraformVariables(dir, []otf.Variable{v1, v2})
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
			rootDir:   "/root",
		}
		want := []string{
			"--ro-bind", "/bins/terraform", "/bin/terraform",
			"--bind", "/root", "/config",
			"--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf",
			"--ro-bind", "/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/certs/ca-certificates.crt",
			"--chdir", "/config",
			"--proc", "/proc",
			"--tmpfs", "/tmp",
			"terraform", "apply",
			"-input=false", "-no-color",
		}
		assert.Equal(t, want, env.buildSandboxArgs([]string{"-input=false", "-no-color"}))
	})

	t.Run("with plugin cache", func(t *testing.T) {
		env := Environment{
			Terraform: &fakeTerraform{"/bins"},
			rootDir:   "/root",
			Config: Config{
				PluginCache: true,
			},
		}
		want := []string{
			"--ro-bind", "/bins/terraform", "/bin/terraform",
			"--bind", "/root", "/config",
			"--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf",
			"--ro-bind", "/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/certs/ca-certificates.crt",
			"--chdir", "/config",
			"--proc", "/proc",
			"--tmpfs", "/tmp",
			"--ro-bind", PluginCacheDir, PluginCacheDir,
			"terraform", "apply",
			"-input=false", "-no-color",
		}
		assert.Equal(t, want, env.buildSandboxArgs([]string{"-input=false", "-no-color"}))
	})

	t.Run("with working directory set", func(t *testing.T) {
		env := Environment{
			Terraform:  &fakeTerraform{"/bins"},
			rootDir:    "/root",
			relWorkDir: "/relative",
			Config: Config{
				PluginCache: true,
			},
		}
		want := []string{
			"--ro-bind", "/bins/terraform", "/bin/terraform",
			"--bind", "/root", "/config",
			"--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf",
			"--ro-bind", "/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/certs/ca-certificates.crt",
			"--chdir", "/config/relative",
			"--proc", "/proc",
			"--tmpfs", "/tmp",
			"--ro-bind", PluginCacheDir, PluginCacheDir,
			"terraform", "apply",
			"-input=false", "-no-color",
		}
		assert.Equal(t, want, env.buildSandboxArgs([]string{"-input=false", "-no-color"}))
	})
}

type fakeEnvironmentApp struct {
	t   *testing.T
	org *otf.Organization
	ws  *otf.Workspace
	otf.Application
}

func (f *fakeEnvironmentApp) GetWorkspace(context.Context, string) (*otf.Workspace, error) {
	return f.ws, nil
}

func (f *fakeEnvironmentApp) CreateRegistrySession(context.Context, string) (otf.RegistrySession, error) {
	return registry.NewTestSession(f.t, f.org), nil
}

func (f *fakeEnvironmentApp) ListVariables(context.Context, string) ([]otf.Variable, error) {
	return nil, nil
}

func (f *fakeEnvironmentApp) Hostname() string { return "fake-host.org" }

func newTestEnvironment(t *testing.T, opts ...otf.NewTestWorkspaceOption) *Environment {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org, opts...)
	cv := otf.NewTestConfigurationVersion(t, ws, otf.ConfigurationVersionCreateOptions{})
	run := otf.NewRun(cv, ws, run.RunCreateOptions{})
	env, err := NewEnvironment(
		context.Background(),
		logr.Discard(),
		&fakeEnvironmentApp{t: t, org: org, ws: ws},
		run,
		nil,
		nil,
		Config{},
	)
	require.NoError(t, err)
	return env
}
