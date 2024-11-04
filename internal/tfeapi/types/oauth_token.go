package types

import (
	"time"

	"github.com/leg100/otf/internal/resource"
)

// OAuthToken represents a VCS configuration including the associated
// OAuth token
type OAuthToken struct {
	ID                  resource.ID `jsonapi:"primary,oauth-tokens"`
	UID                 resource.ID `jsonapi:"attribute" json:"uid"`
	CreatedAt           time.Time   `jsonapi:"attribute" json:"created-at"`
	HasSSHKey           bool        `jsonapi:"attribute" json:"has-ssh-key"`
	ServiceProviderUser string      `jsonapi:"attribute" json:"service-provider-user"`

	// Relations
	OAuthClient *OAuthClient `jsonapi:"relationship" json:"oauth-client"`
}
