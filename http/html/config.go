package html

import (
	"errors"

	"github.com/leg100/otf"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
	oauth2github "golang.org/x/oauth2/github"
	oauth2gitlab "golang.org/x/oauth2/gitlab"
)

var (
	ErrOAuthCredentialsUnspecified = errors.New("no oauth credentials have been specified")
	ErrOAuthCredentialsIncomplete  = errors.New("must specify both client ID and client secret")

	cloudDefaults = map[otf.CloudName]struct {
		hostname string
		scopes   []string
		endpoint oauth2.Endpoint
		cloud    otf.Cloud
	}{
		otf.GithubCloudName: {
			hostname: "github.com",
			endpoint: oauth2github.Endpoint,
			scopes:   []string{"user:email", "read:org"},
			cloud:    otf.GithubCloud{},
		},
		otf.GitlabCloudName: {
			hostname: "gitlab.com",
			endpoint: oauth2gitlab.Endpoint,
			scopes:   []string{"read_user", "read_api"},
			cloud:    otf.GitlabCloud{},
		},
	}
)

// Config is the web app configuration.
type Config struct {
	DevMode bool

	// mapping of cloud name to cloud config
	cloudConfigs map[string]cloudConfig
}

// NewConfigFromFlags binds flags to the config. The flagset must be parsed
// in order for the config to be populated.
func NewConfigFromFlags(flags *pflag.FlagSet) *Config {
	cfg := Config{
		cloudConfigs: make(map[string]cloudConfig, len(cloudDefaults)),
	}

	flags.BoolVar(&cfg.DevMode, "dev-mode", false, "Enable developer mode.")

	for name, defaults := range cloudDefaults {
		cc := cloudConfig{
			hostname: defaults.hostname,
			cloud:    defaults.cloud,
			name:     name,
			endpoint: defaults.endpoint,
			scopes:   defaults.scopes,
		}
		nameStr := string(name)
		flags.StringVar(&cc.clientID, nameStr+"-client-id", "", nameStr+" client ID")
		flags.StringVar(&cc.clientSecret, nameStr+"-client-secret", "", nameStr+" client secret")

		flags.StringVar(&cc.hostname, nameStr+"-hostname", defaults.hostname, nameStr+" hostname")
		flags.BoolVar(&cc.skipTLSVerification, nameStr+"-skip-tls-verification", false, "Skip "+nameStr+" TLS verification")

		cfg.cloudConfigs[nameStr] = cc
	}

	return &cfg
}

// configuration for a cloud provider
type cloudConfig struct {
	hostname            string
	skipTLSVerification bool
	name                otf.CloudName
	cloud               otf.Cloud

	// OAuth config
	clientID     string
	clientSecret string
	scopes       []string
	endpoint     oauth2.Endpoint
}

func (cfg cloudConfig) validate() error {
	if cfg.clientID == "" && cfg.clientSecret == "" {
		return ErrOAuthCredentialsUnspecified
	}
	if cfg.clientID == "" && cfg.clientSecret != "" {
		return ErrOAuthCredentialsIncomplete
	}
	if cfg.clientID != "" && cfg.clientSecret == "" {
		return ErrOAuthCredentialsIncomplete
	}
	return nil
}
