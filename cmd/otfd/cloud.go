package main

import (
	"github.com/leg100/otf"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
)

// cloudConfig bundles together cloud config and oauth config for populating via
// flags
type cloudConfig struct {
	otf.CloudConfig
	*oauth2.Config
}

// newCloudConfigsFromFlags binds flags to cloud configs
func cloudFlags(flags *pflag.FlagSet) []*cloudConfig {
	configs := []*cloudConfig{
		// github
		{
			CloudConfig: otf.GithubDefaults(),
			Config:      otf.GithubOAuthDefaults(),
		},
		// gitlab
		{
			CloudConfig: otf.GitlabDefaults(),
			Config:      otf.GitlabOAuthDefaults(),
		},
	}
	for _, cc := range configs {
		flags.StringVar(&cc.Hostname, cc.Name+"-hostname", cc.Hostname, cc.Name+" hostname")
		flags.BoolVar(&cc.SkipTLSVerification, cc.Name+"-skip-tls-verification", false, "Skip "+cc.Name+" TLS verification")
		flags.StringVar(&cc.ClientID, cc.Name+"-client-id", "", cc.Name+" client ID")
		flags.StringVar(&cc.ClientSecret, cc.Name+"-client-secret", "", cc.Name+" client secret")
	}
	return configs
}
