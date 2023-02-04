package client

import (
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/variable"
)

type (
	stateClient    = state.Client
	variableClient = variable.Client
)

type client struct {
	stateClient
	variableClient
	*http.Client
}

func New(config http.Config) (*client, error) {
	httpClient, err := http.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &client{
		Client:         httpClient,
		stateClient:    stateClient{httpClient},
		variableClient: variableClient{httpClient},
	}, nil
}
