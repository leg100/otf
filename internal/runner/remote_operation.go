package runner

import (
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
)

type RemoteOperationOptions struct {
	Sandbox         bool   // isolate privileged ops within sandbox
	Debug           bool   // toggle debug mode
	PluginCachePath string // toggle use of terraform's shared plugin cache
	TerraformBinDir string // destination directory for terraform binaries

	logger     logr.Logger
	job        *Job
	downloader downloader
	url        string
	jobToken   []byte
	isRemote   bool
}

// newRemoteOperation constructs an operation that is performed remotely,
// communicating via RPC with otfd.
func newRemoteOperation(opts RemoteOperationOptions) (*operation, error) {
	cfg := otfapi.Config{
		URL:           opts.url,
		Token:         string(opts.jobToken),
		RetryRequests: true,
		RetryLogHook: func(_ retryablehttp.Logger, r *http.Request, n int) {
			if n > 0 {
				opts.logger.Error(nil, "retrying request", "url", r.URL, "attempt", n)
			}
		},
	}
	apiClient, err := otfapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
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
