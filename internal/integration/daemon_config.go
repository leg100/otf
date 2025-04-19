package integration

import (
	"time"

	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/runner"
)

// configures the daemon for integration tests
type config struct {
	daemon.Config
	// skip creation of default organization
	skipDefaultOrganization bool
	// github stub server options
	githubOptions []github.TestServerOption
}

type configOption func(*config)

func skipDefaultOrganization() configOption {
	return func(cfg *config) {
		cfg.skipDefaultOrganization = true
	}
}

func withDatabase(dbconn string) configOption {
	return func(cfg *config) {
		cfg.Database = dbconn
	}
}

func withGithubOption(opt github.TestServerOption) configOption {
	return func(cfg *config) {
		cfg.githubOptions = append(cfg.githubOptions, opt)
	}
}

func withGithubOptions(opts ...github.TestServerOption) configOption {
	return func(cfg *config) {
		cfg.githubOptions = opts
	}
}

func withGithubHostname(hostname string) configOption {
	return func(cfg *config) {
		cfg.GithubHostname = hostname
	}
}

func withGithubOAuthCredentials(clientID, clientSecret string) configOption {
	return func(cfg *config) {
		cfg.GithubClientID = clientID
		cfg.GithubClientSecret = clientSecret
	}
}

func withOIDConfig(oidcConfig authenticator.OIDCConfig) configOption {
	return func(cfg *config) {
		cfg.OIDC = oidcConfig
	}
}

func withSiteToken(token string) configOption {
	return func(cfg *config) {
		cfg.SiteToken = token
	}
}

func withSiteAdmins(admins ...string) configOption {
	return func(cfg *config) {
		cfg.SiteAdmins = admins
	}
}

func withTimeouts(planning, applying, checkInterval time.Duration) configOption {
	return func(cfg *config) {
		cfg.PlanningTimeout = planning
		cfg.ApplyingTimeout = applying
		cfg.OverrideTimeoutCheckInterval = checkInterval
	}
}

func disableRunner() configOption {
	return func(cfg *config) {
		cfg.DisableRunner = true
	}
}
func disableScheduler() configOption {
	return func(cfg *config) {
		cfg.DisableScheduler = true
	}
}

type runnerConfigOption func(*runner.Config)

func withTerraformBinDir(dir string) runnerConfigOption {
	return func(cfg *runner.Config) {
		cfg.TerraformBinDir = dir
	}
}

func withRunnerDebug() runnerConfigOption {
	return func(cfg *runner.Config) {
		cfg.Debug = true
	}
}

func withServerRunnerDebug() configOption {
	return func(cfg *config) {
		cfg.RunnerConfig.Debug = true
	}
}

func withServerRunnerSandbox() configOption {
	return func(cfg *config) {
		cfg.RunnerConfig.Sandbox = true
	}
}
