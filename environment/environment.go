package environment

import (
	"context"
	"io"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/workspace"
)

// Environment provides a run with access to various otf services, a working
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
	// RunCLI invokes an arbitrary command
	RunCLI(name string, args ...string) error
	// RunFunc invokes a go func
	RunFunc(fn EnvironmentFunc) error
	// TerraformPath returns the path to the terraform bin
	TerraformPath() string

	Downloader // Downloads TF CLI
	io.Writer  // Permits job to write to output that'll be shown to the user

	// make services available to jobs
	GetWorkspace(ctx context.Context, workspaceID string) (*workspace.Workspace, error)
	ListVariables(ctx context.Context, workspaceID string) ([]*variable.Variable, error)
	GetAgentToken(ctx context.Context, token string) (*auth.AgentToken, error)
	GetPlanFile(ctx context.Context, runID string, format otf.PlanFormat) ([]byte, error)
	UploadPlanFile(ctx context.Context, runID string, plan []byte, format otf.PlanFormat) error
	GetLockFile(ctx context.Context, runID string) ([]byte, error)
	UploadLockFile(ctx context.Context, runID string, lockFile []byte) error
	DownloadConfig(ctx context.Context, cvID string) ([]byte, error)
	otf.StateVersionApp
}

// EnvironmentFunc is a go func invoked within an Environment
type EnvironmentFunc func(context.Context, Environment) error

// Downloader downloads a particular version of software.
type Downloader interface {
	// Download version, return path.
	Download(ctx context.Context, version string, w io.Writer) (path string, err error)
}
