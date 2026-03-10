package agent

import (
	"github.com/leg100/otf/internal/client"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/runner"
)

// New construct an agent runner, i.e. one running remotely to otfd
func New(logger logr.Logger, serverURL string, token string, config *runner.Config) (*runner.Runner, error) {
	config.IsAgent = true
	// Create an API client authenticating with the agent token.
	client, err := client.New(
		logger,
		serverURL,
		token,
	)
	if err != nil {
		return nil, err
	}
	// Construct runner.
	return runner.New(
		logger,
		client,
		// Callback to create a client for a job to interact with otfd.
		func(jobToken string) runner.OperationClient {
			client := client.UseToken(jobToken)
			return client
		},
		config,
	)
}
