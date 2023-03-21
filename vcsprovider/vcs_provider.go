// Package vcsprovider is responsible for VCS providers
package vcsprovider

import (
	"context"
	"time"

	"github.com/leg100/otf/cloud"
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

func (t *VCSProvider) MarshalLog() any {
	return struct {
		ID           string `json:"id"`
		Organization string `json:"organization"`
		Name         string `json:"name"`
		Cloud        string `json:"cloud"`
	}{
		ID:           t.ID,
		Organization: t.Organization,
		Name:         t.Name,
		Cloud:        t.CloudConfig.Name,
	}
}
