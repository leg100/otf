package vcsprovider

import (
	"context"

	"github.com/a-h/templ"
)

type ConfigSchema struct {
	WantsHostname   bool
	DefaultHostname string
	WantsToken      bool
	// TokenDescription is a helpful description of what is expected of the
	// token, e.g. what permissions it should possess.
	TokenDescription string
	// ListInstallations retrieves a list of installations.
	ListInstallations func(context.Context) (ListInstallationsResult, error)
}

type Config struct {
	Token               string
	InstallationID      int64 `schema:"install_id"`
	SkipTLSVerification bool
}

type ListInstallationsResult struct {
	// InstallationLink is a link to the site where a user can create an
	// installation.
	InstallationLink templ.SafeURL
	// Installations is a map of IDs of existing installations keyed by a human
	// meaningful name.
	Installations map[string]int64
}
