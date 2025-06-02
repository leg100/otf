package vcsprovider

import "context"

type ConfigSchema struct {
	WantsHostname   bool
	DefaultHostname string
	WantsToken      bool
	// TokenDescription is a helpful description of what is expected of the
	// token, e.g. what permissions it should possess.
	TokenDescription string
	// ListInstallationIDs retrieves a list of install IDs keyed by a human
	// meaningful name. If non-nil then the retrieved list is presented to the
	// user via the UI in a select input field.
	ListInstallationIDs func(context.Context) (map[string]int64, error)
}

type Config struct {
	Hostname            string
	Token               string
	InstallationID      int64
	SkipTLSVerification bool
}
