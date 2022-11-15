package html

import (
	"errors"
	"fmt"

	"github.com/leg100/otf"
)

var (
	ErrOAuthCredentialsUnspecified = errors.New("no oauth credentials have been specified")
	ErrOAuthCredentialsIncomplete  = errors.New("must specify both client ID and client secret")
)

// Config is the web app configuration.
type Config struct {
	DevMode      bool
	CloudConfigs cloudDB
}

// database of cloud configurations, keyed by name
type cloudDB map[otf.CloudName]*otf.CloudConfig

func (db cloudDB) lookup(name otf.CloudName) (*otf.CloudConfig, error) {
	cfg, ok := db[name]
	if !ok {
		return nil, fmt.Errorf("no such cloud configured: %s", name)
	}
	return cfg, nil
}
