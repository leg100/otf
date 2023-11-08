package agent

import (
	"github.com/spf13/pflag"
)

type (
	// Config is configuration for an agent daemon
	Config struct {
		Organization    *string // only perform operations for a particular org
		Concurrency     int     // number of workers
		Sandbox         bool    // isolate privileged ops within sandbox
		Debug           bool    // toggle debug mode
		PluginCache     bool    // toggle use of terraform's shared plugin cache
		TerraformBinDir string  // destination directory for terraform binaries

		server bool // internal process (true) or separate process (false)
	}
)

func NewConfigFromFlags(flags *pflag.FlagSet) *Config {
	cfg := Config{}
	flags.BoolVar(&cfg.Sandbox, "sandbox", false, "Isolate terraform apply within sandbox for additional security")
	flags.BoolVar(&cfg.Debug, "debug", false, "Enable agent debug mode which dumps additional info to terraform runs.")
	flags.BoolVar(&cfg.PluginCache, "plugin-cache", false, "Enable shared plugin cache for terraform providers.")
	flags.IntVar(&cfg.Concurrency, "concurrency", defaultConcurrency, "Number of runs that can be processed concurrently")
	return &cfg
}
