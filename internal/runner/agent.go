package runner

import (
	apipkg "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/logr"
)

type AgentOptions struct {
	*Config

	Token     string
	ServerURL string
}

func NewAgent(logger logr.Logger, opts AgentOptions) (*Runner, error) {
	// Create an API client authenticating with the agent token.
	client, err := apipkg.NewClient(apipkg.Config{
		URL:           opts.ServerURL,
		Token:         opts.Token,
		Logger:        logger,
		RetryRequests: true,
	})
	if err != nil {
		return nil, err
	}
	opts.isAgent = true

	// Construct runner.
	return New(
		logger,
		&Client{Client: client},
		func(jobToken []byte) (OperationClient, error) {
			return NewRemoteOperationClient(jobToken, opts.ServerURL, logger)
		},
		*opts.Config,
	)
}
