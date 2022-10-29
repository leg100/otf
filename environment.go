package otf

import (
	"context"
	"io"
)

// Environment provides a Job with access to various oTF services, a working
// directory, and the ability to invoke arbitrary commands and go functions.
// Invoking commands and functions via the environment means the environment can
// handle canceling them if necessary.
type Environment interface {
	Path() string
	RunTerraform(cmd string, args ...string) error
	RunCLI(name string, args ...string) error
	RunFunc(fn EnvironmentFunc) error
	TerraformPath() string

	// All app services should be made available to the environment
	Application
	// For downloading TF CLI
	Downloader
	// Permits job to write to output that'll be shown to the user
	io.Writer
}

// EnvironmentFunc is a go func that is invoked within an environment (and with
// access to the environment).
type EnvironmentFunc func(context.Context, Environment) error

// Downloader is capable of downloading a version of software.
type Downloader interface {
	// Download version, return path.
	Download(ctx context.Context, version string, w io.Writer) (path string, err error)
}
