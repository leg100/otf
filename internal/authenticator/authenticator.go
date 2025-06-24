// Package authenticator is responsible for handling the authentication of users with
// third party identity providers.
package authenticator

import (
	"github.com/leg100/otf/internal/user"
)

// UserInfo is info about a user retrieved from the identity provider.
type UserInfo struct {
	Username  user.Username
	AvatarURL *string // Optional.
}
