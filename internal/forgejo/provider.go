package forgejo

import (
	"context"

	"github.com/leg100/otf/internal/vcs"
)

func RegisterVCSKind(vcsService *vcs.Service, hostname string, skipTLSVerification bool) {
	vcsService.RegisterKind(vcs.Kind{
		ID:   vcs.KindID("forgejo"),
		Name: "Forgejo",
		Icon: Icon(),
		TokenKind: &vcs.TokenKind{
			Description: tokenDescription(hostname),
		},
		Hostname:     hostname,
		EventHandler: HandleEvent,
		NewClient: func(ctx context.Context, cfg vcs.Config) (vcs.Client, error) {
			return NewTokenClient(vcs.NewTokenClientOptions{
				Token:               *cfg.Token,
				Hostname:            hostname,
				SkipTLSVerification: skipTLSVerification,
			})
		},
	})
}
