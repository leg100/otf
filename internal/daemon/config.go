package daemon

import (
	"errors"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/engine"
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
	AllowedOrigins               string
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

	tokens.GoogleIAPConfig
}

// NewConfig constructs an otfd configuration with defaults.
func NewConfig() Config {
	return Config{
		RunnerConfig:  runner.NewConfig(),
		CacheConfig:   &inmem.CacheConfig{},
		MaxConfigSize: configversion.DefaultConfigMaxSize,
		DefaultEngine: engine.Default,
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
