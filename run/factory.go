package run

import (
	"context"

	"github.com/leg100/otf/configversion"
)

// factory constructs runs
type factory struct {
	ConfigurationVersionService
	WorkspaceService
}

// NewRun constructs a new run at the beginning of its lifecycle using the
// provided options.
func (f *factory) NewRun(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error) {
	ws, err := f.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var cv *configversion.ConfigurationVersion
	if opts.ConfigurationVersionID != nil {
		cv, err = f.GetConfigurationVersion(ctx, *opts.ConfigurationVersionID)
	} else {
		cv, err = f.GetLatestConfigurationVersion(ctx, workspaceID)
	}
	if err != nil {
		return nil, err
	}

	return NewRun(cv, ws, opts), nil
}
