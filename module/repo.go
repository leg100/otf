package module

import (
	"strings"

	"github.com/leg100/otf"
)

// Repo is the path of repository for a module. It is expected to follow a
// certain format, which varies according to the cloud providing the Repo, but
// it should always end with '/<identifier>-<name>-<provider>', with name and
// provider being used to set the name and provider of the module.
type Repo string

func (r Repo) Split() (name, provider string, err error) {
	repoParts := strings.Split(string(r), "/")
	if len(repoParts) < 2 {
		return "", "", otf.ErrInvalidRepo
	}
	parts := strings.SplitN(repoParts[1], "-", 3)
	if len(parts) < 3 {
		return "", "", ErrInvalidModuleRepo
	}
	return parts[len(parts)-1], parts[len(parts)-2], nil
}
