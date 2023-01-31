package otf

import (
	"context"
	"io"
)

// Environment provides a Job with access to various otf services, a working
// directory, and the ability to invoke arbitrary commands and go functions.
// Invoking commands and functions via the environment means the environment can
// handle canceling them if necessary.
type Environment interface {
	// Path returns absolute root path
	Path() string
	// WorkingDir returns relative path for terraform operations
	WorkingDir() string
	// RunTerraform invokes a terraform command
	RunTerraform(cmd string, args ...string) error
	// RunCLI runs an arbitrary command
	RunCLI(name string, args ...string) error
	// RunFunc runs a go func with access to the env
	RunFunc(fn EnvironmentFunc) error
	// TerraformPath is the path to the terraform bin
	TerraformPath() string

	Client     // All client services should be made available to the environment
	Downloader // For downloading TF CLI
	io.Writer  // Permits job to write to output that'll be shown to the user
}

// EnvironmentFunc is a go func that is invoked within an environment (and with
// access to the environment).
type EnvironmentFunc func(context.Context, Environment) error

// Downloader is capable of downloading a version of software.
type Downloader interface {
	// Download version, return path.
	Download(ctx context.Context, version string, w io.Writer) (path string, err error)
}
