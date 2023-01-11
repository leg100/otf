package main

import (
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/github"
	"github.com/leg100/otf/gitlab"
	"github.com/spf13/pflag"
)

// newCloudConfigsFromFlags binds flags to cloud configs
func cloudFlags(flags *pflag.FlagSet) cloud.OAuthConfigs {
	configs := cloud.OAuthConfigs{
		// github
		{
			Config:      github.Defaults(),
			OAuthConfig: github.OAuthDefaults(),
		},
		// gitlab
		{
			Config:      gitlab.Defaults(),
			OAuthConfig: gitlab.OAuthDefaults(),
		},
	}
	for _, cc := range configs {
		flags.StringVar(&cc.Hostname, cc.Name+"-hostname", cc.Hostname, cc.Name+" hostname")
		flags.BoolVar(&cc.SkipTLSVerification, cc.Name+"-skip-tls-verification", false, "Skip "+cc.Name+" TLS verification")
		flags.StringVar(&cc.OAuthConfig.ClientID, cc.Name+"-client-id", "", cc.Name+" client ID")
		flags.StringVar(&cc.OAuthConfig.ClientSecret, cc.Name+"-client-secret", "", cc.Name+" client secret")
	}
	return configs
}
