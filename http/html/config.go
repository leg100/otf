package html

import (
	"fmt"
	"net/url"
)

const (
	githubCallbackPath = "/github/callback"
)

// Config is the web app configuration.
type Config struct {
	GithubClientID     string
	GithubClientSecret string
	GithubRedirectURL  string
	DevMode            bool
}

func (cfg *Config) validate() error {
	u, err := url.Parse(cfg.GithubRedirectURL)
	if err != nil {
		return fmt.Errorf("invalid github redirect URL: %w", err)
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	}

	if u.Scheme != "https" {
		return fmt.Errorf("github redirect URL scheme must set to https")
	}

	if u.Path == "" {
		u.Path = githubCallbackPath
	}

	if u.Path != githubCallbackPath {
		return fmt.Errorf("github redirect URL path must be set to %s", githubCallbackPath)
	}

	return nil
}
