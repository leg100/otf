package html

import (
	"fmt"
	"net/url"
	"strings"
)

// Config is the web app configuration.
type Config struct {
	Github GithubConfig

	DevMode bool
}

func (cfg *Config) validate() error {
	if !strings.HasPrefix(cfg.Github.RedirectURL, "https://") {
		return fmt.Errorf("github redirect url must start with https://")
	}

	u, err := url.Parse(cfg.Github.RedirectURL)
	if err != nil {
		return fmt.Errorf("invalid github redirect URL: %w", err)
	}

	if u.Path != githubCallbackPath {
		return fmt.Errorf("github redirect URL path must be set to %s", githubCallbackPath)
	}

	return nil
}
