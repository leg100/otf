package otf

import (
	"context"

	"golang.org/x/oauth2"
)

type CloudConfig interface {
	NewClient(context.Context, CloudClientOptions) (CloudClient, error)
	Valid() error
	Name() string
	Hostname() string
	Endpoint() (oauth2.Endpoint, error)
	Scopes() []string
	ClientID() string
	ClientSecret() string
	SkipTLSVerification() bool
}

type CloudConfigMixin struct {
	*OAuthCredentials

	cloudName           string
	hostname            string
	endpoint            oauth2.Endpoint
	scopes              []string
	skipTLSVerification bool
}

type cloudConfigOptions struct {
	hostname            *string
	skipTLSVerification *bool
}

func (g CloudConfigMixin) CloudName() string         { return g.cloudName }
func (g CloudConfigMixin) Hostname() string          { return g.hostname }
func (g CloudConfigMixin) Scopes() []string          { return g.scopes }
func (g CloudConfigMixin) SkipTLSVerification() bool { return g.skipTLSVerification }

func (g CloudConfigMixin) Endpoint() (oauth2.Endpoint, error) {
	var err error
	var endpoint oauth2.Endpoint

	endpoint.AuthURL, err = UpdateHost(g.endpoint.AuthURL, g.hostname)
	if err != nil {
		return oauth2.Endpoint{}, err
	}
	endpoint.TokenURL, err = UpdateHost(g.endpoint.TokenURL, g.hostname)
	if err != nil {
		return oauth2.Endpoint{}, err
	}
	return endpoint, nil
}

func (g *CloudConfigMixin) override(opts *cloudConfigOptions) {
	if opts == nil {
		return
	}
	if opts.hostname != nil {
		g.hostname = *opts.hostname
	}
	if opts.skipTLSVerification != nil {
		g.skipTLSVerification = *opts.skipTLSVerification
	}
}
