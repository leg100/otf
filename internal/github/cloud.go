package github

import (
	"context"
	"net/http"

	"github.com/leg100/otf/internal/cloud"
)

type Cloud struct{}

func (g *Cloud) NewClient(ctx context.Context, opts cloud.ClientOptions) (cloud.Client, error) {
	return NewClient(ctx, opts)
}

func (Cloud) HandleEvent(w http.ResponseWriter, r *http.Request, secret string) *cloud.VCSEvent {
	return HandleEvent(w, r, secret)
}
