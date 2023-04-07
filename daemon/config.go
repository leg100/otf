package daemon

import (
	"reflect"

	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/github"
	"github.com/leg100/otf/gitlab"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/tokens"
)

// Config configures the otfd daemon. Descriptions of each field can be found in
// the flag definitions in ./cmd/otfd
type Config struct {
	AgentConfig                  *agent.Config
	CacheConfig                  *inmem.CacheConfig
	Github                       cloud.CloudOAuthConfig
	Gitlab                       cloud.CloudOAuthConfig
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
