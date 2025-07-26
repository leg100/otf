package gitlab

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/vcs"
)

const KindID vcs.KindID = "gitlab"

func RegisterVCSKind(vcsService *vcs.Service, apiURL *internal.WebURL, skipTLSVerification bool) {
	vcsService.RegisterKind(vcs.Kind{
		ID:   KindID,
		Icon: Icon(),
		TokenKind: &vcs.TokenKind{
			Description: tokenDescription(apiURL.Host),
		},
		DefaultURL:   apiURL,
		EventHandler: HandleEvent,
		NewClient: func(ctx context.Context, cfg vcs.ClientConfig) (vcs.Client, error) {
			return NewTokenClient(vcs.NewTokenClientOptions{
				Token:               *cfg.Token,
				BaseURL:             apiURL,
				SkipTLSVerification: skipTLSVerification,
			})
		},
		TFEServiceProviders: []vcs.TFEServiceProviderType{
			vcs.ServiceProviderGitlab,
			vcs.ServiceProviderGitlabCE,
			vcs.ServiceProviderGitlabEE,
		},
	})
}
