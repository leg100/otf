package daemon

import (
	"errors"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/configversion"
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
	GithubHostname               string
	GithubClientID               string
	GithubClientSecret           string
	GitlabHostname               string
	GitlabClientID               string
	GitlabClientSecret           string
	OIDC                         authenticator.OIDCConfig
	Secret                       []byte // 16-byte secret for signing URLs and encrypting payloads
	SiteToken                    string
	Host                         string
	WebhookHost                  string
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
	SkipTLSVerification          bool
	// skip checks for latest terraform version
	DisableLatestChecker         *bool
	PlanningTimeout              time.Duration
	ApplyingTimeout              time.Duration
	OverrideTimeoutCheckInterval time.Duration

	tokens.GoogleIAPConfig
}

func ApplyDefaults(cfg *Config) {
	if cfg.RunnerConfig == nil {
		cfg.RunnerConfig = &runner.Config{
			MaxJobs: runner.DefaultMaxJobs,
		}
	}
	if cfg.CacheConfig == nil {
		cfg.CacheConfig = &inmem.CacheConfig{}
	}
	if cfg.MaxConfigSize == 0 {
		cfg.MaxConfigSize = configversion.DefaultConfigMaxSize
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
