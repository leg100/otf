package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSandboxArgs(t *testing.T) {
	t.Run("without plugin cache", func(t *testing.T) {
		env := execution{
			workdir: &workdir{root: "/root"},
		}
		want := []string{
			"bwrap",
			"--ro-bind", "/tmp/tf-bins/1.1.1/terraform", "/bin/terraform",
			"--bind", "/root", "/config",
			"--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf",
			"--ro-bind", "/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/certs/ca-certificates.crt",
			"--chdir", "/config",
			"--proc", "/proc",
			"--tmpfs", "/tmp",
			"terraform", "apply",
			"-input=false", "-no-color",
		}
		assert.Equal(t, want, env.addSandboxWrapper([]string{"/tmp/tf-bins/1.1.1/terraform", "apply", "-input=false", "-no-color"}))
	})

	t.Run("with plugin cache", func(t *testing.T) {
		env := execution{
			Config: Config{
				PluginCache: true,
			},
			workdir: &workdir{root: "/root"},
		}
		want := []string{
			"bwrap",
			"--ro-bind", "/tmp/tf-bins/1.1.1/terraform", "/bin/terraform",
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
		assert.Equal(t, want, env.addSandboxWrapper([]string{"/tmp/tf-bins/1.1.1/terraform", "apply", "-input=false", "-no-color"}))
	})

	t.Run("with relative working directory", func(t *testing.T) {
		env := execution{
			Config: Config{
				PluginCache: true,
			},
			workdir: &workdir{root: "/root", relative: "/relative"},
		}
		want := []string{
			"bwrap",
			"--ro-bind", "/tmp/tf-bins/1.1.1/terraform", "/bin/terraform",
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
		assert.Equal(t, want, env.addSandboxWrapper([]string{"/tmp/tf-bins/1.1.1/terraform", "apply", "-input=false", "-no-color"}))
	})
}
