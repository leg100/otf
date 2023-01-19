package agent

import (
	"github.com/spf13/pflag"
)

// Config configures the agent.
type Config struct {
	Organization *string // only process runs belonging to org
	External     bool    // dedicated agent (true) or otfd (false)
	Concurrency  int     // number of workers
	Sandbox      bool    // isolate privileged ops within sandbox
	Debug        bool    // toggle debug mode
	PluginCache  bool    // toggle use of terraform's shared plugin cache
}

func NewConfigFromFlags(flags *pflag.FlagSet) *Config {
	cfg := Config{}
	flags.BoolVar(&cfg.Sandbox, "sandbox", false, "Isolate terraform apply within sandbox for additional security")
	flags.BoolVar(&cfg.Debug, "debug", false, "Enable agent debug mode which dumps additional info to terraform runs.")
	flags.BoolVar(&cfg.PluginCache, "plugin-cache", false, "Enable shared plugin cache for terraform providers.")
	flags.IntVar(&cfg.Concurrency, "concurrency", DefaultConcurrency, "Number of runs that can be processed concurrently")
	return &cfg
}
