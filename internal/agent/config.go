package agent

import (
	"github.com/leg100/otf/internal/api"
	"github.com/spf13/pflag"
)

type (
	// Config is configuration for an agent.
	Config struct {
		Organization    *string // only process runs belonging to org
		External        bool    // dedicated agent (true) or integrated into otfd (false)
		Concurrency     int     // number of workers
		Sandbox         bool    // isolate privileged ops within sandbox
		Debug           bool    // toggle debug mode
		PluginCache     bool    // toggle use of terraform's shared plugin cache
		TerraformBinDir string  // destination directory for terraform binaries
	}
	// ExternalConfig is configuration for an external agent
	ExternalConfig struct {
		APIConfig api.Config

		Config
	}
)

func NewConfigFromFlags(flags *pflag.FlagSet) *Config {
	cfg := Config{}
	flags.BoolVar(&cfg.Sandbox, "sandbox", false, "Isolate terraform apply within sandbox for additional security")
	flags.BoolVar(&cfg.Debug, "debug", false, "Enable agent debug mode which dumps additional info to terraform runs.")
	flags.BoolVar(&cfg.PluginCache, "plugin-cache", false, "Enable shared plugin cache for terraform providers.")
	flags.IntVar(&cfg.Concurrency, "concurrency", DefaultConcurrency, "Number of runs that can be processed concurrently")
	return &cfg
}

func NewExternalConfigFromFlags(flags *pflag.FlagSet) *ExternalConfig {
	cfg := ExternalConfig{
		Config: *NewConfigFromFlags(flags),
	}
	flags.StringVar(&cfg.APIConfig.Address, "address", api.DefaultAddress, "Address of OTF server")
	flags.StringVar(&cfg.APIConfig.Token, "token", "", "Agent token for authentication")
	return &cfg
}
