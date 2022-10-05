package html

import (
	"context"

	"github.com/leg100/otf"
	"golang.org/x/oauth2"
)

type CloudConfig interface {
	Valid() error
	NewCloud() (Cloud, error)
}

type Cloud interface {
	CloudName() string
	Endpoint() oauth2.Endpoint
	Scopes() []string
	ClientID() string
	ClientSecret() string
	NewDirectoryClient(context.Context, DirectoryClientOptions) (DirectoryClient, error)
}

type DirectoryClientOptions struct {
	Token *oauth2.Token

	*oauth2.Config
}

type DirectoryClient interface {
	GetUser(ctx context.Context) (string, error)
	ListOrganizations(ctx context.Context) ([]string, error)
}

func updateEndpoint(endpoint oauth2.Endpoint, hostname string) (oauth2.Endpoint, error) {
	var err error

	endpoint.AuthURL, err = otf.UpdateHost(endpoint.AuthURL, hostname)
	if err != nil {
		return oauth2.Endpoint{}, err
	}
	endpoint.TokenURL, err = otf.UpdateHost(endpoint.TokenURL, hostname)
	if err != nil {
		return oauth2.Endpoint{}, err
	}
	return endpoint, nil
}
