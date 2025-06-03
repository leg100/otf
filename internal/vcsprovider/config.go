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
	Token               *string
	Installation        *Installation
	SkipTLSVerification bool
}

type ListInstallationsResult struct {
	// InstallationLink is a link to the site where a user can create an
	// installation.
	InstallationLink templ.SafeURL
	// Results is a map of IDs of existing installations keyed by a human
	// meaningful name.
	Results []Installation
}

type Installation struct {
	ID           int64
	AppID        int64
	Username     *string
	Organization *string
}

func (i Installation) String() string {
	if i.Organization != nil {
		return "org/" + *i.Organization
	}
	return "user/" + *i.Username
}
