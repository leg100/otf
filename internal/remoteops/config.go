package remoteops

import (
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/spf13/pflag"
)

type (
	// Config is configuration for a remote operations daemon.
	Config struct {
		Organization    *string // only perform operations for a particular org
		Concurrency     int     // number of workers
		Sandbox         bool    // isolate privileged ops within sandbox
		Debug           bool    // toggle debug mode
		PluginCache     bool    // toggle use of terraform's shared plugin cache
		TerraformBinDir string  // destination directory for terraform binaries

		isAgent bool // external agent process (true) or integrated into otfd (false)
	}
	// AgentConfig is configuration for an agent, i.e. a remote operations
	// daemon that communicates with the server via RPC.
	AgentConfig struct {
		APIConfig otfapi.Config

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

func NewAgentConfigFromFlags(flags *pflag.FlagSet) *AgentConfig {
	cfg := AgentConfig{
		Config: *NewConfigFromFlags(flags),
	}
	flags.StringVar(&cfg.APIConfig.Address, "address", otfapi.DefaultAddress, "Address of OTF server")
	flags.StringVar(&cfg.APIConfig.Token, "token", "", "Agent token for authentication")
	return &cfg
}
