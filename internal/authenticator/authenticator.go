// Package authenticator is responsible for handling the authentication of users with
// third party identity providers.
package authenticator

import (
	"net/url"

	"github.com/leg100/otf/internal/user"
)

type UserInfo struct {
	Username  user.Username
	AvatarURL *url.URL
}
