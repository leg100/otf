package runner

import (
	"github.com/leg100/otf/internal/engine"
	"github.com/spf13/pflag"
)

type Config struct {
	Name         string // descriptive name given to runner
	MaxJobs      int    // number of jobs the runner can execute at any one time
	LauncherKind string // how jobs are launched: forked processes or kubernetes jobs
	OperationConfig
}

func NewDefaultConfig() *Config {
	return &Config{
		MaxJobs:      DefaultMaxJobs,
		LauncherKind: "process",
	}
}

func NewConfigFromFlags(flags *pflag.FlagSet) *Config {
	opts := Config{}
	flags.IntVar(&opts.MaxJobs, "concurrency", DefaultMaxJobs, "Number of runs that can be processed concurrently")
	flags.BoolVar(&opts.Sandbox, "sandbox", false, "Isolate terraform apply within sandbox for additional security")
	flags.BoolVar(&opts.Debug, "debug", false, "Enable runner debug mode which dumps additional info to terraform runs.")
	flags.BoolVar(&opts.PluginCache, "plugin-cache", false, "Enable shared plugin cache for terraform providers.")
	flags.StringVar(&opts.engineBinDir, "engine-bins-dir", engine.DefaultBinDir, "Destination directory for engine binary downloads.")
	flags.StringVar(&opts.LauncherKind, "launcher", "process", "Sets how jobs are launched: process or kubernetes")
	return &opts
}
