package github

import (
	"context"
	"net/http"

	"github.com/leg100/otf"
)

type Cloud struct{}

func (g *Cloud) NewClient(ctx context.Context, opts otf.CloudClientOptions) (otf.CloudClient, error) {
	return NewClient(ctx, opts)
}

func (Cloud) HandleEvent(w http.ResponseWriter, r *http.Request, opts otf.HandleEventOptions) *otf.VCSEvent {
	return nil
}
