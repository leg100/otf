package runner

import (
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/logr"
	"github.com/spf13/pflag"
)

type Config struct {
	OperationConfig

	Name         string       // descriptive name given to runner
	MaxJobs      int          // number of jobs the runner can execute at any one time
	ExecutorKind ExecutorKind // how jobs are launched: forked processes or kubernetes jobs
	KubeConfig   KubeConfig
	LoggerConfig *logr.Config
}

func NewDefaultConfig() *Config {
	return &Config{
		MaxJobs:      DefaultMaxJobs,
		ExecutorKind: processExecutorKind,
	}
}

func NewConfigFromFlags(flags *pflag.FlagSet, loggerConfig *logr.Config) *Config {
	opts := Config{
		LoggerConfig: loggerConfig,
	}
	flags.IntVar(&opts.MaxJobs, "concurrency", DefaultMaxJobs, "Number of runs that can be processed concurrently")
	flags.BoolVar(&opts.Debug, "debug", false, "Enable runner debug mode which dumps additional info to terraform runs.")
	flags.BoolVar(&opts.PluginCache, "plugin-cache", false, "Enable shared plugin cache for terraform providers.")
	flags.StringVar(&opts.EngineBinDir, "engine-bins-dir", engine.DefaultBinDir, "Destination directory for engine binary downloads.")
	flags.Var(&opts.ExecutorKind, "executor", "Executor for executing jobs: 'process' or 'kubernetes'")
	return &opts
}
