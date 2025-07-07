package daemon

import (
	"encoding/hex"
	"errors"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/inmem"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/tokens"
)

var ErrInvalidSecretLength = errors.New("secret must be 16 bytes in size")

// Config configures the otfd daemon. Descriptions of each field can be found in
// the flag definitions in ./cmd/otfd
type Config struct {
	RunnerConfig       *runner.Config
	CacheConfig        *inmem.CacheConfig
	GithubHostname     string
	GithubClientID     string
	GithubClientSecret string
	GitlabHostname     string
	GitlabClientID     string
	GitlabClientSecret string
	ForgejoHostname    string // TODO: forgejo is often self-hosted, and there may be more than one of them.  this should be a per-VCS setting
	OIDC               authenticator.OIDCConfig
	Secret             Secret `name:"secret" help:"Hex-encoded 16 byte secret for cryptographic work. Required." required:""`
	SiteToken          string `name:"site-token" help:"API token with site-wide unlimited permissions. Use with care."`
	Host               string `name:"hostname" help:"User-facing hostname for otf"`
	WebhookHost        string `name:"webhook-hostname" help:"External hostname for otf webhooks"`
	// TODO: rename --address to --listen
	Address                      string `help:"Listening address" default:":8080"`
	Database                     string `help:"Postgres connection string" default:"postgres:///otf?host=/var/run/postgresql"`
	AllowedOrigins               string `name:"allowed-origins" help:"Allowed origins for websocket upgrades"`
	MaxConfigSize                int64  `name:"max-config-size" default:"${max_config_size}" help:"Maximum permitted configuration size in bytes."`
	SSL                          bool   `name:"ssl" help:"Toggle SSL"`
	CertFile, KeyFile            string
	EnableRequestLogging         bool
	DisableScheduler             bool
	DisableRunner                bool
	RestrictOrganizationCreation bool
	SiteAdmins                   []string `name:"site-admins" help:"Promote a list of users to site admin."`
	SkipTLSVerification          bool
	// skip checks for latest terraform version
	DisableLatestChecker         *bool
	PlanningTimeout              time.Duration
	ApplyingTimeout              time.Duration
	OverrideTimeoutCheckInterval time.Duration
	DefaultEngine                *engine.Engine
	LogConfig                    logr.Config

	tokens.GoogleIAPConfig
}

// 16-byte secret for signing URLs and encrypting payloads
type Secret []byte

func (id *Secret) UnmarshalText(text []byte) error {
	*id = make([]byte, 16)
	n, err := hex.Decode(*id, text)
	if err != nil {
		return err
	}
	if n != 16 {
		return ErrInvalidSecretLength
	}
	return nil
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
