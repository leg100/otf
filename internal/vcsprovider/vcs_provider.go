// Package vcsprovider is responsible for VCS providers
package vcsprovider

import (
	"context"
	"time"

	"github.com/leg100/otf/internal/cloud"
	"golang.org/x/exp/slog"
)

type (
	// VCSProvider provides authenticated access to a VCS. Equivalent to an OAuthClient in
	// TFE.
	VCSProvider struct {
		ID           string
		CreatedAt    time.Time
		Name         string       // TODO: rename to description (?)
		CloudConfig  cloud.Config // cloud config for creating client
		Token        string       // credential for creating client
		Organization string       // vcs provider belongs to an organization
	}
)

func (t *VCSProvider) String() string { return t.Name }

func (t *VCSProvider) NewClient(ctx context.Context) (cloud.Client, error) {
	return t.CloudConfig.NewClient(ctx, cloud.Credentials{
		PersonalToken: &t.Token,
	})
}

// LogValue implements slog.LogValuer.
func (t *VCSProvider) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", t.ID),
		slog.String("organization", t.Organization),
		slog.String("name", t.Name),
		slog.String("cloud", t.CloudConfig.Name),
	)
}
