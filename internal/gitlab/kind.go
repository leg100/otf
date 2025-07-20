package gitlab

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/vcs"
)

const KindID vcs.KindID = "gitlab"

func RegisterVCSKind(vcsService *vcs.Service, baseURL *internal.URL, skipTLSVerification bool) {
	vcsService.RegisterKind(vcs.Kind{
		ID:   KindID,
		Icon: Icon(),
		TokenKind: &vcs.TokenKind{
			Description: tokenDescription(baseURL.Host),
		},
		BaseURL:      baseURL,
		EventHandler: HandleEvent,
		NewClient: func(ctx context.Context, cfg vcs.Config) (vcs.Client, error) {
			return NewTokenClient(vcs.NewTokenClientOptions{
				Token:               *cfg.Token,
				BaseURL:             baseURL,
				SkipTLSVerification: skipTLSVerification,
			})
		},
		TFEServiceProvider: vcs.ServiceProviderGitlab,
	})
}
