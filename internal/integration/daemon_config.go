package integration

import (
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/runner"
)

// configures the daemon for integration tests
type config struct {
	daemon.Config
	// skip creation of default organization
	skipDefaultOrganization bool
	// skip setting up an automatic github server stub
	skipGithubStub bool
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
		cfg.GithubHostname = internal.MustWebURL(hostname)
		// setting a hostname implies the test is setting up its own stub so
		// skip setting up another stub
		cfg.skipGithubStub = true
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

func withDeleteRunsAfter(deleteRunsAfter, checkInterval time.Duration) configOption {
	return func(cfg *config) {
		cfg.DeleteRunsAfter = deleteRunsAfter
		cfg.OverrideDeleterInterval = checkInterval
	}
}

func withDeleteConfigsAfter(deleteConfigsAfter, checkInterval time.Duration) configOption {
	return func(cfg *config) {
		cfg.DeleteConfigsAfter = deleteConfigsAfter
		cfg.OverrideDeleterInterval = checkInterval
	}
}

func withKubernetesExecutor(configPath, image, namespace string, serverURL runner.KubeConfigServerURL, ttl time.Duration) configOption {
	return func(cfg *config) {
		cfg.RunnerConfig.ExecutorKind = runner.KubeExecutorKind
		cfg.RunnerConfig.KubeConfig.ConfigPath = configPath
		cfg.RunnerConfig.KubeConfig.Image = image
		cfg.RunnerConfig.KubeConfig.Namespace = namespace
		cfg.RunnerConfig.KubeConfig.ServerURL = serverURL
		cfg.RunnerConfig.KubeConfig.TTLAfterFinish = ttl
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

func withDefaultEngine(engine *engine.Engine) configOption {
	return func(cfg *config) {
		cfg.DefaultEngine = engine
	}
}

func withKeyPairPaths(private, public string) configOption {
	return func(cfg *config) {
		cfg.PrivateKeyPath = private
		cfg.PublicKeyPath = public
	}
}

func withHostname(hostname string) configOption {
	return func(cfg *config) {
		cfg.Host = hostname
	}
}

func withSSLDisabled() configOption {
	return func(cfg *config) {
		cfg.SSL = false
	}
}

type agentConfigOption func(*runner.AgentOptions)

func withEngineBinDir(dir string) agentConfigOption {
	return func(opts *runner.AgentOptions) {
		opts.EngineBinDir = dir
	}
}

func withRunnerDebug() agentConfigOption {
	return func(opts *runner.AgentOptions) {
		opts.Debug = true
	}
}

// withAgentURL sets the URL of the server that the agent talks to.
func withAgentURL(url string) agentConfigOption {
	return func(opts *runner.AgentOptions) {
		opts.ServerURL = url
	}
}
