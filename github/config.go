package github

import (
	"github.com/leg100/otf"
	"github.com/spf13/pflag"
)

type Config struct {
	otf.CloudConfigMixin
}

func NewConfigFromFlags(flags *pflag.FlagSet) *Config {
	cfg := defaults()

	flags.StringVar(&cfg.hostname, "github-hostname", cfg.hostname, "Github hostname")
	flags.BoolVar(&cfg.skipTLSVerification, "github-skip-tls-verification", false, "Skip github TLS verification")
	cfg.OAuthCredentials.AddFlags(flags)

	return cfg
}

func (cfg *Config) NewCloud() (Cloud, error) {
	return &Cloud{Config: cfg}, nil
}

func defaults() *Config {
	return &Config{
		CloudConfigMixin: otf.CloudConfigMixin{
			OAuthCredentials: &otf.OAuthCredentials{prefix: "github"},
			cloudName:        "github",
			endpoint:         oauth2github.Endpoint,
			scopes:           []string{"user:email", "read:org"},
			hostname:         DefaultGithubHostname,
		},
	}
}
