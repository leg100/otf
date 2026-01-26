package runner

import (
	"github.com/spf13/pflag"
)

type Config struct {
	OperationConfig

	Name         string       // descriptive name given to runner
	MaxJobs      int          // number of jobs the runner can execute at any one time. Only applicable to the 'fork' excecutor.
	ExecutorKind ExecutorKind // how jobs are launched: forked processes or kubernetes jobs
	KubeConfig   *kubeConfig
}

func NewDefaultConfig() *Config {
	return &Config{
		MaxJobs:         DefaultMaxJobs,
		ExecutorKind:    ForkExecutorKind,
		KubeConfig:      defaultKubeConfig,
		OperationConfig: defaultOperationConfig(),
	}
}

func RegisterFlags(flags *pflag.FlagSet, cfg *Config) {
	flags.IntVar(&cfg.MaxJobs, "concurrency", cfg.MaxJobs, "Number of runs that can be processed concurrently. Only applicable to the fork executor.")
	flags.Var(&cfg.ExecutorKind, "executor", "Executor for executing jobs: 'fork' or 'kubernetes'")
	RegisterOperationFlags(flags, &cfg.OperationConfig)
	registerKubeFlags(flags, cfg.KubeConfig)
}
