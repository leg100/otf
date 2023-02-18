// Package client allows remote interaction with the otf application
package client

import (
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/watch"
)

type (
	stateClient    = state.Client
	variableClient = variable.Client
	authClient     = auth.Client
	watchClient    = watch.Client
)

type client struct {
	*http.Client
	http.Config

	stateClient
	variableClient
	authClient
	watchClient
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
		authClient:     authClient{httpClient},
		watchClient:    watchClient{config},
	}, nil
}
