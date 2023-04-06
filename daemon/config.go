package daemon

import (
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/cloud"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/github"
	"github.com/leg100/otf/gitlab"
	"github.com/leg100/otf/inmem"
)

// Config configures the otfd daemon. Descriptions of each field can be found in
// the flag definitions in ./cmd/otfd
type Config struct {
	AgentConfig                  *agent.Config
	LoggerConfig                 *cmdutil.LoggerConfig
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
	DisableRunScheduler          bool
	RestrictOrganizationCreation bool

	auth.GoogleIAPConfig
}

func NewDefaultConfig() Config {
	return Config{
		AgentConfig: &agent.Config{
			Concurrency: agent.DefaultConcurrency,
		},
		CacheConfig: &inmem.CacheConfig{},
		Github: cloud.CloudOAuthConfig{
			Config:      github.Defaults(),
			OAuthConfig: github.OAuthDefaults(),
		},
		Gitlab: cloud.CloudOAuthConfig{
			Config:      gitlab.Defaults(),
			OAuthConfig: gitlab.OAuthDefaults(),
		},
	}
}
