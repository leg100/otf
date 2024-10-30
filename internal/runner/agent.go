package runner

import (
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/releases"
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
		RetryRequests: true,
	})
	if err != nil {
		return nil, err
	}
	return newRunner(
		logger,
		&remoteClient{Client: apiClient},
		&remoteOperationSpawner{
			logger:     logger,
			url:        opts.URL,
			downloader: releases.NewDownloader(opts.TerraformBinDir),
		},
		true,
		*opts.Config,
	)
}

type remoteOperationSpawner struct {
	logger     logr.Logger
	downloader downloader
	url        string
}

func (r *remoteOperationSpawner) newOperation(job *Job, jobToken []byte) (*operation, error) {
	apiClient, err := otfapi.NewClient(otfapi.Config{
		URL:           r.url,
		Token:         string(jobToken),
		RetryRequests: true,
	})
	if err != nil {
		return nil, err
	}
	return newOperation(operationOptions{
		logger:     r.logger,
		job:        job,
		jobToken:   jobToken,
		downloader: r.downloader,
		runs:       &run.Client{Client: apiClient},
		jobs:       &remoteClient{Client: apiClient},
		workspaces: &workspace.Client{Client: apiClient},
		variables:  &variable.Client{Client: apiClient},
		state:      &state.Client{Client: apiClient},
		configs:    &configversion.Client{Client: apiClient},
		logs:       &logs.Client{Client: apiClient},
		server:     apiClient,
		isRemote:   true,
	}), nil
}
