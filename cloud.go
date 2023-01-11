package otf

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/leg100/otf/cloud"
	"golang.org/x/oauth2"
)

// Cloud is an external provider of various cloud services e.g. identity provider, VCS
// repositories etc.
type Cloud interface {
	NewClient(context.Context, CloudClientOptions) (cloud.Client, error)
	EventHandler
}

// CloudConfig is configuration for a cloud provider
type CloudConfig struct {
	Name                string
	Hostname            string
	SkipTLSVerification bool

	Cloud
}

// CloudOAuthConfig is the configuration for a cloud provider and its OAuth
// configuration.
type CloudOAuthConfig struct {
	CloudConfig
	*oauth2.Config
}

type CloudOAuthConfigs []*CloudOAuthConfig

// CloudConfigs returns the list of cloud configs from a list of cloud oauth
// configs
func (c CloudOAuthConfigs) CloudConfigs() []CloudConfig {
	var configs []CloudConfig
	for _, cc := range c {
		configs = append(configs, cc.CloudConfig)
	}
	return configs
}

// CloudClientOptions are options for constructing a cloud client
type CloudClientOptions struct {
	Hostname            string
	SkipTLSVerification bool

	CloudCredentials
}

// CloudCredentials are credentials for a cloud client
type CloudCredentials struct {
	// tokens are mutually-exclusive - at least one must be specified
	OAuthToken    *oauth2.Token
	PersonalToken *string
}

type CloudService interface {
	GetCloudConfig(name string) (CloudConfig, error)
	ListCloudConfigs() []CloudConfig
}

func (cfg CloudConfig) String() string {
	return string(cfg.Name)
}

func (cfg *CloudConfig) NewClient(ctx context.Context, creds CloudCredentials) (cloud.Client, error) {
	return cfg.Cloud.NewClient(ctx, CloudClientOptions{
		Hostname:            cfg.Hostname,
		SkipTLSVerification: cfg.SkipTLSVerification,
		CloudCredentials:    creds,
	})
}

func (cfg *CloudConfig) HTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.SkipTLSVerification,
			},
		},
	}
}
