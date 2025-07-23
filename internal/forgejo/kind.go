package forgejo

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/vcs"
)

func RegisterVCSKind(vcsService *vcs.Service, apiURL *internal.WebURL, skipTLSVerification bool) {
	vcsService.RegisterKind(vcs.Kind{
		ID:   vcs.KindID("forgejo"),
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
			vcs.ServiceProviderForgejo,
		},
	})
}
