// Package gitlab provides gitlab related code
package gitlab

import (
	"github.com/leg100/otf/internal/vcs"
	oauth2gitlab "golang.org/x/oauth2/gitlab"
)

const (
	DefaultHostname string = "gitlab.com"
)

var (
	OAuthEndpoint = oauth2gitlab.Endpoint
	OAuthScopes   = []string{"read_user", "read_api"}
)

func init() {
	vcs.RegisterPersonalTokenClientConstructor(vcs.GitlabKind, NewPersonalTokenClient)
}
