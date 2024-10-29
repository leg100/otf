package operation

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
)

type LocalOperationOptions struct {
	Sandbox         bool   // isolate privileged ops within sandbox
	Debug           bool   // toggle debug mode
	PluginCachePath string // toggle use of terraform's shared plugin cache
	TerraformBinDir string // destination directory for terraform binaries

	logger     logr.Logger
	job        *Job
	downloader downloader
	url        string
	jobToken   []byte
}

// newLocalOperation constructs an operation that is performed remotely,
// communicating via RPC with otfd.
func newLocalOperation(ctx context.Context, opts LocalOperationOptions) (*operation, error) {
	// this is a server agent: directly authenticate as job with services
	ctx = internal.AddSubjectToContext(ctx, opts.job)

	return newOperation(operationOptions{
		runs:       &run.Client{Client: apiClient},
		workspaces: &workspace.Client{Client: apiClient},
		variables:  &variable.Client{Client: apiClient},
		runners:    &client{Client: apiClient},
		state:      &state.Client{Client: apiClient},
		configs:    &configversion.Client{Client: apiClient},
		logs:       &logs.Client{Client: apiClient},
		jobToken:   opts.jobToken,
		server:     apiClient,
	}), nil
}
