package runner

import (
	"github.com/leg100/otf/internal/engine"
	"github.com/spf13/pflag"
)

type Config struct {
	Name         string // descriptive name given to runner
	MaxJobs      int    // number of jobs the runner can execute at any one time
	Sandbox      bool   // isolate privileged ops within sandbox
	Debug        bool   // toggle debug mode
	PluginCache  bool   // toggle use of terraform's shared plugin cache
	EngineBinDir string // destination directory for engine binaries
}

func NewConfig() *Config {
	return &Config{
		MaxJobs: DefaultMaxJobs,
	}
}

func NewConfigFromFlags(flags *pflag.FlagSet) *Config {
	opts := Config{}
	flags.IntVar(&opts.MaxJobs, "concurrency", DefaultMaxJobs, "Number of runs that can be processed concurrently")
	flags.BoolVar(&opts.Sandbox, "sandbox", false, "Isolate terraform apply within sandbox for additional security")
	flags.BoolVar(&opts.Debug, "debug", false, "Enable runner debug mode which dumps additional info to terraform runs.")
	flags.BoolVar(&opts.PluginCache, "plugin-cache", false, "Enable shared plugin cache for terraform providers.")
	flags.StringVar(&opts.EngineBinDir, "engine-bins-dir", engine.DefaultBinDir, "Destination directory for engine binary downloads.")
	return &opts
}
