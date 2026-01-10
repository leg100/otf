package runner

import (
	"github.com/leg100/otf/internal/engine"
	"github.com/spf13/pflag"
)

type Config struct {
	OperationConfig

	Name         string       // descriptive name given to runner
	MaxJobs      int          // number of jobs the runner can execute at any one time
	ExecutorKind ExecutorKind // how jobs are launched: forked processes or kubernetes jobs
	KubeConfig   *kubeConfig
}

func NewDefaultConfig() *Config {
	return &Config{
		MaxJobs:      DefaultMaxJobs,
		ExecutorKind: processExecutorKind,
		KubeConfig:   defaultKubeConfig,
	}
}

func RegisterFlags(flags *pflag.FlagSet, cfg *Config, agent bool) {
	flags.IntVar(&cfg.MaxJobs, "concurrency", DefaultMaxJobs, "Number of runs that can be processed concurrently")
	flags.BoolVar(&cfg.Debug, "debug", false, "Enable runner debug mode which dumps additional info to terraform runs.")
	flags.BoolVar(&cfg.PluginCache, "plugin-cache", false, "Enable shared plugin cache for terraform providers.")
	flags.StringVar(&cfg.EngineBinDir, "engine-bins-dir", engine.DefaultBinDir, "Destination directory for engine binary downloads.")
	flags.Var(&cfg.ExecutorKind, "executor", "Executor for executing jobs: 'process' or 'kubernetes'")
	registerKubeFlags(flags, cfg.KubeConfig, agent)
}
