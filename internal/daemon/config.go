package daemon

import (
	"reflect"

	"github.com/leg100/otf/internal/agent"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/gitlab"
	"github.com/leg100/otf/internal/inmem"
	"github.com/leg100/otf/internal/tokens"
)

// Config configures the otfd daemon. Descriptions of each field can be found in
// the flag definitions in ./cmd/otfd
type Config struct {
	AgentConfig                  *agent.Config
	CacheConfig                  *inmem.CacheConfig
	Github                       cloud.CloudOAuthConfig
	Gitlab                       cloud.CloudOAuthConfig
	OIDC                         cloud.OIDCConfig
	Secret                       string // secret for signing URLs
	SiteToken                    string
	Host                         string
	Address                      string
	Database                     string
	MaxConfigSize                int64
	SSL                          bool
	CertFile, KeyFile            string
	EnableRequestLogging         bool
	DevMode                      bool
	DisableScheduler             bool
	RestrictOrganizationCreation bool
	SiteAdmins                   []string

	tokens.GoogleIAPConfig
}

func ApplyDefaults(cfg *Config) {
	if cfg.AgentConfig == nil {
		cfg.AgentConfig = &agent.Config{
			Concurrency: agent.DefaultConcurrency,
		}
	}
	if cfg.CacheConfig == nil {
		cfg.CacheConfig = &inmem.CacheConfig{}
	}
	if cfg.MaxConfigSize == 0 {
		cfg.MaxConfigSize = configversion.DefaultConfigMaxSize
	}
	if reflect.ValueOf(cfg.Github).IsZero() {
		cfg.Github = cloud.CloudOAuthConfig{
			Config:      github.Defaults(),
			OAuthConfig: github.OAuthDefaults(),
		}
	}
	if reflect.ValueOf(cfg.Gitlab).IsZero() {
		cfg.Gitlab = cloud.CloudOAuthConfig{
			Config:      gitlab.Defaults(),
			OAuthConfig: gitlab.OAuthDefaults(),
		}
	}
}
