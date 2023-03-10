package otf

import (
	"errors"
	"strings"
)

var ErrInvalidModuleRepo = errors.New("invalid repository name for module")

// ModuleRepo is the path of repository for a module. It is expected to follow a
// certain format, which varies according to the cloud providing the ModuleRepo, but
// it should always end with '/<identifier>-<name>-<provider>', with name and
// provider being used to set the name and provider of the module.
type ModuleRepo string

func (r ModuleRepo) Split() (name, provider string, err error) {
	repoParts := strings.Split(string(r), "/")
	if len(repoParts) < 2 {
		return "", "", ErrInvalidRepo
	}
	parts := strings.SplitN(name, "-", 3)
	if len(parts) < 3 {
		return "", "", ErrInvalidModuleRepo
	}
	return parts[len(parts)-2], parts[len(parts)-1], nil
}
