package runner

import (
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
	"github.com/spf13/pflag"
)

type AgentOptions struct {
	*Config

	URL   string
	Token string
}

func NewAgentOptionsFromFlags(flags *pflag.FlagSet) *AgentOptions {
	opts := AgentOptions{
		Config: NewConfigFromFlags(flags),
	}
	flags.StringVar(&opts.Name, "name", "", "Give agent a descriptive name. Optional.")
	flags.StringVar(&opts.URL, "url", otfapi.DefaultURL, "URL of OTF server")
	flags.StringVar(&opts.Token, "token", "", "Agent token for authentication")
	return &opts
}

func NewAgent(logger logr.Logger, opts AgentOptions) (*Runner, error) {
	apiClient, err := otfapi.NewClient(otfapi.Config{
		URL:           opts.URL,
		Token:         opts.Token,
		Logger:        logger,
		RetryRequests: true,
	})
	if err != nil {
		return nil, err
	}
	return newRunner(
		logger,
		&remoteClient{Client: apiClient},
		&remoteOperationSpawner{
			logger: logger,
			config: *opts.Config,
			url:    opts.URL,
		},
		true,
		*opts.Config,
	)
}

type remoteOperationSpawner struct {
	config Config
	logger logr.Logger
	url    string
}

func (s *remoteOperationSpawner) newOperation(job *Job, jobToken []byte) (*operation, error) {
	apiClient, err := otfapi.NewClient(otfapi.Config{
		URL:           s.url,
		Token:         string(jobToken),
		Logger:        s.logger,
		RetryRequests: true,
	})
	if err != nil {
		return nil, err
	}
	return newOperation(operationOptions{
		logger:       s.logger,
		Debug:        s.config.Debug,
		Sandbox:      s.config.Sandbox,
		PluginCache:  s.config.PluginCache,
		engineBinDir: s.config.EngineBinDir,
		job:          job,
		jobToken:     jobToken,
		runs:         &run.Client{Client: apiClient},
		jobs:         newRemoteClient(apiClient, nil),
		workspaces:   &workspace.Client{Client: apiClient},
		variables:    &variable.Client{Client: apiClient},
		state:        &state.Client{Client: apiClient},
		configs:      &configversion.Client{Client: apiClient},
		server:       apiClient,
		isAgent:      true,
	}), nil
}
