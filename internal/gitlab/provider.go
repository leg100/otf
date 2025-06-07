package gitlab

import (
	"context"

	"github.com/leg100/otf/internal/vcs"
)

const KindID vcs.KindID = "gitlab"

type Provider struct {
	Hostname            string
	SkipTLSVerification bool
}

func (p *Provider) Register(vcsService *vcs.Service) {
	vcsService.RegisterKind(vcs.Kind{
		ID:   KindID,
		Name: "GitLab",
		Icon: Icon(),
		TokenKind: &vcs.TokenKind{
			Description: tokenDescription(p.Hostname),
		},
		Hostname: p.Hostname,
		NewClient: func(ctx context.Context, cfg vcs.Config) (vcs.Client, error) {
			return NewTokenClient(vcs.NewTokenClientOptions{
				Token:               *cfg.Token,
				Hostname:            p.Hostname,
				SkipTLSVerification: p.SkipTLSVerification,
			})
		},
	})
}
