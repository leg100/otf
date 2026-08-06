package runner

import (
	"fmt"
	"strconv"

	"github.com/spf13/pflag"
)

type Config struct {
	OperationConfig

	Name         string       // descriptive name given to runner
	MaxJobs      *int         // number of jobs the runner can execute at any one time.
	ExecutorKind ExecutorKind // how jobs are launched: forked processes or kubernetes jobs
	KubeConfig   kubeConfig
}

func NewDefaultConfig() *Config {
	return &Config{
		ExecutorKind:    ForkExecutorKind,
		KubeConfig:      defaultKubeConfig,
		OperationConfig: defaultOperationConfig(),
	}
}

func (c *Config) resolveMaxJobs() (int, error) {
	if c.MaxJobs == nil {
		if c.ExecutorKind == ForkExecutorKind {
			return DefaultMaxJobs, nil
		}
		// Unlimited.
		return 0, nil
	}
	switch {
	case *c.MaxJobs < 0:
		return 0, fmt.Errorf("concurrency must not be negative: %d", *c.MaxJobs)
	case *c.MaxJobs == 0 && c.ExecutorKind == ForkExecutorKind:
		// Forking an unlimited number of processes would exhaust the host.
		return 0, fmt.Errorf("concurrency must be at least 1 for the %s executor", ForkExecutorKind)
	}
	return *c.MaxJobs, nil
}

type ConcurrencyValue struct{ p **int }

var _ pflag.Value = ConcurrencyValue{}

func (ConcurrencyValue) Type() string { return "int" }

func (c ConcurrencyValue) String() string {
	if c.p == nil || *c.p == nil {
		return ""
	}
	return strconv.Itoa(**c.p)
}

func (c ConcurrencyValue) Set(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*c.p = &n
	return nil
}

func RegisterFlags(flags *pflag.FlagSet, cfg *Config) {
	flags.Var(ConcurrencyValue{p: &cfg.MaxJobs}, "concurrency", fmt.Sprintf("Number of runs that can be processed concurrently. Defaults to %d for the %s executor, and to unlimited (0) for the %s executor.", DefaultMaxJobs, ForkExecutorKind, KubeExecutorKind))
	flags.Var(&cfg.ExecutorKind, "executor", "Executor for executing jobs: 'fork' or 'kubernetes'")
	RegisterOperationFlags(flags, &cfg.OperationConfig)
	registerKubeFlags(flags, &cfg.KubeConfig)
}
