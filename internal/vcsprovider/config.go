package vcsprovider

import (
	"context"
	"errors"
	"fmt"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal"
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

func (c Config) validate() error {
	// Either token or installation must be set but not both
	if c.Token == nil && c.Installation == nil {
		return errors.New("must set one of token or installation")
	}
	if c.Token != nil && c.Installation != nil {
		return errors.New("cannot set both a token and an installation")
	}
	// If token is set it cannot be empty
	if c.Token != nil && *c.Token == "" {
		return fmt.Errorf("token: %w", internal.ErrEmptyValue)
	}
	return nil
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

func (i Installation) validate() error {
	// IDs cannot be zero
	if i.ID == 0 {
		return errors.New("install ID cannot be zero")
	}
	if i.AppID == 0 {
		return errors.New("install app ID cannot be zero")
	}
	// Either token or installation must be set but not both
	if i.Username == nil && i.Organization == nil {
		return errors.New("must set one of token or installation")
	}
	if i.Username != nil && i.Organization != nil {
		return errors.New("cannot set both a token and an installation")
	}
	// Neither username nor organization, if set, can be empty
	if i.Username != nil && *i.Username == "" {
		return fmt.Errorf("install username: %w", internal.ErrEmptyValue)
	}
	if i.Organization != nil && *i.Organization == "" {
		return fmt.Errorf("install organization: %w", internal.ErrEmptyValue)
	}
	return nil
}
