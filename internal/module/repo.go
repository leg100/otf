package module

import (
	"strings"

	"github.com/leg100/otf/internal/vcs"
)

// Repo is the path of repository for a module. It is expected to follow a
// certain format, which varies according to the cloud providing the Repo, but
// it should always end with '/<identifier>-<name>-<provider>', with name and
// provider being used to set the name and provider of the module.
type Repo vcs.Repo

func (r Repo) Split() (name, provider string, err error) {
	parts := strings.SplitN(r.Name, "-", 3)
	if len(parts) < 3 {
		return "", "", ErrInvalidModuleRepo
	}
	return parts[len(parts)-1], parts[len(parts)-2], nil
}
