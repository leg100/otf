package otf

import (
	"golang.org/x/oauth2"
)

type CloudConfig interface {
	Valid() error
	NewCloud() (Cloud, error)
	CloudName() string
	Hostname() string
	Endpoint() (oauth2.Endpoint, error)
	Scopes() []string
	ClientID() string
	ClientSecret() string
	SkipTLSVerification() bool
}

// mixin for CloudConfig implementations
type cloudConfig struct {
	*OAuthCredentials

	cloudName           string
	hostname            string
	endpoint            oauth2.Endpoint
	scopes              []string
	skipTLSVerification bool
}

// optional overrides
type cloudConfigOptions struct {
	hostname            *string
	skipTLSVerification *bool
}

func (g cloudConfig) CloudName() string         { return g.cloudName }
func (g cloudConfig) Hostname() string          { return g.hostname }
func (g cloudConfig) Scopes() []string          { return g.scopes }
func (g cloudConfig) SkipTLSVerification() bool { return g.skipTLSVerification }

func (g cloudConfig) Endpoint() (oauth2.Endpoint, error) {
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

func (g *cloudConfig) override(opts *cloudConfigOptions) {
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
