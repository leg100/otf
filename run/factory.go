package run

import (
	"context"

	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/workspace"
)

// factory constructs runs
type factory struct {
	config    configversion.Service
	workspace workspace.Service
}

// NewRun constructs a new run at the beginning of its lifecycle using the
// provided options.
func (f *factory) NewRun(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error) {
	ws, err := f.workspace.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var cv *configversion.ConfigurationVersion
	if opts.ConfigurationVersionID != nil {
		cv, err = f.config.GetConfigurationVersion(ctx, *opts.ConfigurationVersionID)
	} else {
		cv, err = f.config.GetLatestConfigurationVersion(ctx, workspaceID)
	}
	if err != nil {
		return nil, err
	}

	return NewRun(cv, ws, opts), nil
}
