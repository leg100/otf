package cloud

import (
	"context"
	"crypto/tls"
	"net/http"

	"golang.org/x/oauth2"
)

// Config is configuration for a cloud provider
type Config struct {
	Name                string
	Hostname            string
	SkipTLSVerification bool

	Cloud
}

func (cfg Config) String() string {
	return string(cfg.Name)
}

func (cfg *Config) NewClient(ctx context.Context, creds Credentials) (Client, error) {
	return cfg.Cloud.NewClient(ctx, ClientOptions{
		Hostname:            cfg.Hostname,
		SkipTLSVerification: cfg.SkipTLSVerification,
		Credentials:         creds,
	})
}

func (cfg *Config) HTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.SkipTLSVerification,
			},
		},
	}
}

// Credentials are credentials for a cloud client
type Credentials struct {
	// tokens are mutually-exclusive - at least one must be specified
	OAuthToken    *oauth2.Token
	PersonalToken *string
}

// CloudOAuthConfig is the configuration for a cloud provider and its OAuth
// configuration.
type CloudOAuthConfig struct {
	Config
	OAuthConfig *oauth2.Config
}

type OAuthConfigs []*CloudOAuthConfig

// Configs returns the list of cloud configs from a list of cloud oauth
// configs
func (cfgs OAuthConfigs) Configs() []Config {
	var configs []Config
	for _, cfg := range cfgs {
		configs = append(configs, cfg.Config)
	}
	return configs
}
