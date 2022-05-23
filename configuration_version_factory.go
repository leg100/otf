package otf

import (
	"context"
)

// ConfigurationVersionFactory creates ConfigurationVersion objects
type ConfigurationVersionFactory struct {
	WorkspaceService WorkspaceService
}

// NewConfigurationVersion creates a ConfigurationVersion object from scratch
func (f *ConfigurationVersionFactory) NewConfigurationVersion(workspaceID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error) {
	cv := ConfigurationVersion{
		ID:            NewID("cv"),
		autoQueueRuns: DefaultAutoQueueRuns,
		status:        ConfigurationPending,
		source:        DefaultConfigurationSource,
	}

	if opts.AutoQueueRuns != nil {
		cv.autoQueueRuns = *opts.AutoQueueRuns
	}

	if opts.Speculative != nil {
		cv.speculative = *opts.Speculative
	}

	ws, err := f.WorkspaceService.Get(context.Background(), WorkspaceSpec{ID: &workspaceID})
	if err != nil {
		return nil, err
	}
	cv.Workspace = &Workspace{ID: ws.ID}

	return &cv, nil
}

// NewConfigurationVersionFromDefaults creates a new run with defaults.
func NewConfigurationVersionFromDefaults(ws *Workspace) *ConfigurationVersion {
	cv := ConfigurationVersion{
		ID:     NewID("cv"),
		status: ConfigurationPending,
	}
	cv.Workspace = &Workspace{ID: ws.ID}
	return &cv
}
