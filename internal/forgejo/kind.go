package forgejo

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/vcs"
)

func RegisterVCSKind(vcsService *vcs.Service, baseURL *internal.URL, skipTLSVerification bool) {
	vcsService.RegisterKind(vcs.Kind{
		ID:   vcs.KindID("forgejo"),
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
	})
}
