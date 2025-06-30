package runner

import (
	"context"

	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
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
	opts.OperationConfig.IsAgent = true
	return newRunner(
		logger,
		&remoteClient{Client: apiClient},
		&RemoteOperationSpawner{
			Logger: logger,
			Config: opts.OperationConfig,
			URL:    opts.URL,
		},
		true,
		*opts.Config,
	)
}

type RemoteOperationSpawner struct {
	Config OperationConfig
	Logger logr.Logger
	URL    string
}

func (s *RemoteOperationSpawner) NewOperation(ctx context.Context, jobID resource.TfeID, jobToken []byte) (*operation, error) {
	client, err := otfapi.NewClient(otfapi.Config{
		URL:           s.URL,
		Token:         string(jobToken),
		Logger:        s.logger,
		RetryRequests: true,
	})
	if err != nil {
		return nil, err
	}
	return newOperation(ctx, operationOptions{
		logger:          s.Logger,
		OperationConfig: s.Config,
		jobID:           jobID,
		jobToken:        jobToken,
		runs:            &run.Client{Client: client},
		jobs:            &remoteClient{Client: client},
		workspaces:      &workspace.Client{Client: client},
		variables:       &variable.Client{Client: client},
		state:           &state.Client{Client: client},
		configs:         &configversion.Client{Client: client},
		server:          client,
	})
}
