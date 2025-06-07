package forgejo

import (
	"context"

	"github.com/leg100/otf/internal/vcs"
)

const Kind vcs.Kind = "forgejo"

type Provider struct {
	Hostname            string
	SkipTLSVerification bool
}

func (p *Provider) Register(vcsService *vcs.Service) {
	vcsService.RegisterSchema(Kind, vcs.ProviderKind{
		Kind: Kind,
		Name: "Forgejo",
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
