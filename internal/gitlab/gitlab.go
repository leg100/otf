// Package gitlab provides gitlab related code
package gitlab

import (
	"github.com/leg100/otf/internal"
	oauth2gitlab "golang.org/x/oauth2/gitlab"
)

var (
	DefaultBaseURL = internal.MustURL("https://gitlab.com/api/v4")
	OAuthEndpoint  = oauth2gitlab.Endpoint
	OAuthScopes    = []string{"read_user", "read_api"}
)
