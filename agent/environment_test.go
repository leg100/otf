package agent

import (
	"path"
	"testing"

	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
)

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
			env := newTestEnvironment(t, &workspace.Workspace{WorkingDirectory: tt.workdir})
			assert.Equal(t, tt.workdir, env.relWorkDir)
			assert.DirExists(t, env.rootDir)
			assert.DirExists(t, env.absWorkDir)
			assert.FileExists(t, path.Join(env.absWorkDir, "terraform.tfvars"))
		})
	}
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
