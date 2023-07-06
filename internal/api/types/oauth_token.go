package types

import "time"

// OAuthToken represents a VCS configuration including the associated
// OAuth token
type OAuthToken struct {
	ID                  string    `jsonapi:"primary,oauth-tokens"`
	UID                 string    `jsonapi:"attribute" json:"uid"`
	CreatedAt           time.Time `jsonapi:"attribute" json:"created-at"`
	HasSSHKey           bool      `jsonapi:"attribute" json:"has-ssh-key"`
	ServiceProviderUser string    `jsonapi:"attribute" json:"service-provider-user"`

	// Relations
	OAuthClient *OAuthClient `jsonapi:"relationship" json:"oauth-client"`
}
