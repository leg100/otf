package otf

import "context"

// Environment provides a Job with access to various oTF services, a working
// directory, and the ability to invoke arbitrary commands and go functions.
// Invoking commands and functions via the environment means the environment can
// handle canceling them if necessary.
type Environment interface {
	Path() string
	ConfigurationVersionService() ConfigurationVersionService
	StateVersionService() StateVersionService
	RunService() RunService
	RunCLI(name string, args ...string) error
	RunFunc(fn EnvironmentFunc) error
}

// EnvironmentFunc is a go func that is invoked within an environment (and with
// access to the environment).
type EnvironmentFunc func(context.Context, Environment) error
