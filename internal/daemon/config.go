package daemon

import (
	"errors"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/forgejo"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/gitlab"
	"github.com/leg100/otf/internal/inmem"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/tokens"
)

var ErrInvalidSecretLength = errors.New("secret must be 16 bytes in size")

// Config configures the otfd daemon. Descriptions of each field can be found in
// the flag definitions in ./cmd/otfd
type Config struct {
	RunnerConfig                 *runner.Config
	CacheConfig                  *inmem.CacheConfig
	GithubHostname               *internal.WebURL
	GithubClientID               string
	GithubClientSecret           string
	GitlabHostname               *internal.WebURL
	GitlabClientID               string
	GitlabClientSecret           string
	ForgejoHostname              *internal.WebURL // TODO: forgejo is often self-hosted, and there may be more than one of them.  this should be a per-VCS setting
	OIDC                         authenticator.OIDCConfig
	Secret                       []byte // 16-byte secret for signing URLs and encrypting payloads
	PublicKeyPath                string
	PrivateKeyPath               string
	SiteToken                    string
	Host                         string
	WebhookHost                  string
	Address                      string
	Database                     string
	MaxConfigSize                int64
	SSL                          bool
	CertFile, KeyFile            string
	EnableRequestLogging         bool
	DisableScheduler             bool
	DisableRunner                bool
	RestrictOrganizationCreation bool
	SiteAdmins                   []string
	SkipTLSVerification          bool
	// skip checks for latest terraform version
	DisableLatestChecker         *bool
	PlanningTimeout              time.Duration
	ApplyingTimeout              time.Duration
	OverrideTimeoutCheckInterval time.Duration
	DefaultEngine                *engine.Engine
	DeleteRunsAfter              time.Duration
	DeleteConfigsAfter           time.Duration
	OverrideDeleterInterval      time.Duration

	tokens.GoogleIAPConfig
}

// NewConfig constructs an otfd configuration with defaults.
func NewConfig() Config {
	return Config{
		RunnerConfig:    runner.NewDefaultConfig(),
		CacheConfig:     &inmem.CacheConfig{},
		MaxConfigSize:   configversion.DefaultConfigMaxSize,
		DefaultEngine:   engine.Default,
		GithubHostname:  github.DefaultBaseURL(),
		GitlabHostname:  gitlab.DefaultBaseURL,
		ForgejoHostname: forgejo.DefaultBaseURL,
	}
}

func (cfg *Config) Valid() error {
	if cfg.Secret == nil {
		return &internal.ErrMissingParameter{Parameter: "secret"}
	}
	if len(cfg.Secret) != 16 {
		return ErrInvalidSecretLength
	}
	return nil
}
